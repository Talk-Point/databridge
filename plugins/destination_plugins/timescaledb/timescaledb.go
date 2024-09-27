package timescaledb

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/Talk-Point/databridge/models"
	"github.com/Talk-Point/databridge/plugins"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

type TimescaleDBDestination struct {
	Model     *models.Model
	DB        *sql.DB
	Table     string
	Schema    map[string]string // Column types
	BatchSize int
}

func (d *TimescaleDBDestination) Init(config map[string]interface{}, model *models.Model) error {
	d.Model = model

	connStr := os.Getenv("TIMESCALEDB_CONN_STR")
	if connStr == "" {
		return fmt.Errorf("TIMESCALEDB_CONN_STR environment variable is required")
	}
	table := config["table"].(string)
	d.Table = table

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	d.DB = db

	// Optionally, create table if not exists using d.Schema
	return nil
}

func (d *TimescaleDBDestination) Close() error {
	return d.DB.Close()
}

func (d *TimescaleDBDestination) getSQLType(columnType models.ColumnType) string {
	switch columnType {
	case models.String:
		return "TEXT"
	case models.BigInt:
		return "BIGINT"
	case models.Float:
		return "NUMERIC(10,4)"
	case models.DateTime:
		return "TIMESTAMPTZ NOT NULL"
	case models.DateTimeNullable:
		return "TIMESTAMPTZ"
	case models.Int:
		return "INTEGER"
	default:
		return "TEXT"
	}
}

func (d *TimescaleDBDestination) CreateSchema() ([]string, error) {
	queries := []string{}

	var stm strings.Builder

	stm.WriteString(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n", d.Table))

	for i, column := range d.Model.Columns {
		stm.WriteString(fmt.Sprintf("    %s %s", column.Name, d.getSQLType(column.Type)))
		if i < len(d.Model.Columns)-1 {
			stm.WriteString(",\n")
		} else {
			stm.WriteString("\n")
		}
	}

	if len(d.Model.Unique) > 0 {
		stm.WriteString(",\n    PRIMARY KEY (")
		for i, key := range d.Model.Unique {
			stm.WriteString(key)
			if i < len(d.Model.Unique)-1 {
				stm.WriteString(", ")
			}
		}
		stm.WriteString(")\n")
	}

	stm.WriteString(");\n")

	queries = append(queries, stm.String())

	// check if in d.Model.Unique the time column
	var hasTimeColumn bool
	for _, key := range d.Model.Unique {
		if key == "time" {
			hasTimeColumn = true
			break
		}
	}
	if hasTimeColumn {
		queries = append(queries, fmt.Sprintf("SELECT create_hypertable('%s', 'time', if_not_exists => TRUE);", d.Table))
	}

	return queries, nil
}

func (d *TimescaleDBDestination) RunSchema() error {
	queries, err := d.CreateSchema()
	if err != nil {
		return err
	}

	connStr := os.Getenv("TIMESCALEDB_CONN_STR")
	if !strings.Contains(connStr, "sslmode") {
		connStr += "?sslmode=disable"
	}
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %v", err)
	}
	defer db.Close()

	for _, query := range queries {
		log.WithFields(log.Fields{
			"query": query,
		}).Debug("Running schema queries")
		_, err := db.Exec(query)
		if err != nil {
			return fmt.Errorf("error running schema query: %v", err)
		}
	}

	return nil
}

func (d *TimescaleDBDestination) InsertQuery() string {
	var stm strings.Builder

	stm.WriteString(fmt.Sprintf("INSERT INTO %s (", d.Table))
	for i, column := range d.Model.Columns {
		stm.WriteString(column.Name)
		if i < len(d.Model.Columns)-1 {
			stm.WriteString(", ")
		}
	}
	stm.WriteString(") VALUES (")

	valueIndex := 1
	for i, _ := range d.Model.Columns {
		stm.WriteString(fmt.Sprintf("$%d", valueIndex))
		valueIndex++
		if i < len(d.Model.Columns)-1 {
			stm.WriteString(", ")
		}
	}
	stm.WriteString(") ")

	// Add ON CONFLICT clause
	stm.WriteString(" ON CONFLICT (")
	for i, uniqueColumn := range d.Model.Unique {
		stm.WriteString(uniqueColumn)
		if i < len(d.Model.Unique)-1 {
			stm.WriteString(", ")
		}
	}
	stm.WriteString(") DO UPDATE SET ")
	for i, column := range d.Model.Columns {
		if contains(d.Model.Unique, column.Name) {
			continue
		}
		stm.WriteString(fmt.Sprintf("%s = EXCLUDED.%s", column.Name, column.Name))
		if i < len(d.Model.Columns)-1 {
			stm.WriteString(", ")
		}
	}
	stm.WriteString(";")
	return stm.String()
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (d *TimescaleDBDestination) StoreData(data []map[string]interface{}) (int, int, error) {
	return d.StoreDataBatch(data)
}

func (d *TimescaleDBDestination) StoreDataSingle(data []map[string]interface{}) (int, int, error) {
	connStr := os.Getenv("TIMESCALEDB_CONN_STR")
	if !strings.Contains(connStr, "sslmode") {
		connStr += "?sslmode=disable"
	}
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return 0, 0, fmt.Errorf("unable to connect to database: %v", err)
	}
	defer db.Close()

	q := d.InsertQuery()
	log.WithFields(log.Fields{
		"query":        q,
		"destionation": "timescaledb",
	}).Debug("insert query")

	total_success := 0
	total_failed := 0
	for idx, record := range data {
		values := make([]interface{}, len(d.Model.Columns))
		for i, column := range d.Model.Columns {
			if record[column.Name] == "" {
				values[i] = nil
				continue
			}
			values[i] = record[column.Name]
		}

		_, err := db.Exec(q, values...)
		if err != nil {
			log.WithFields(log.Fields{
				"idx":    idx,
				"record": record,
				"type":   "failed",
				"error":  err,
			}).Error("inserting record")
			total_failed++
		} else {
			log.WithFields(log.Fields{
				"idx":    idx,
				"record": record,
				"type":   "success",
			}).Debug("inserting record")
			total_success++
		}
	}

	return total_success, total_failed, nil
}

func (d *TimescaleDBDestination) StoreDataBatch(data []map[string]interface{}) (int, int, error) {
	connStr := os.Getenv("TIMESCALEDB_CONN_STR")
	if !strings.Contains(connStr, "sslmode") {
		connStr += "?sslmode=disable"
	}
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return 0, 0, fmt.Errorf("unable to connect to database: %v", err)
	}
	defer db.Close()

	q := d.InsertQuery()
	log.WithFields(log.Fields{
		"query":        q,
		"destionation": "timescaledb",
	}).Debug("insert query")

	totalSuccess := 0
	totalFailed := 0
	batchSize := d.BatchSize
	batch := make([][]interface{}, 0, batchSize)

	executeBatch := func() error {
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		stmt, err := tx.Prepare(q)
		if err != nil {
			tx.Rollback()
			return err
		}
		defer stmt.Close()

		for _, values := range batch {
			_, err := stmt.Exec(values...)
			if err != nil {
				tx.Rollback()
				return err
			}
		}
		err = tx.Commit()
		if err != nil {
			return err
		}
		return nil
	}

	for idx, record := range data {
		values := make([]interface{}, len(d.Model.Columns))
		for i, column := range d.Model.Columns {
			if record[column.Name] == "" {
				values[i] = nil
				continue
			}
			values[i] = record[column.Name]
		}
		batch = append(batch, values)

		if len(batch) == batchSize {
			err := executeBatch()
			if err != nil {
				log.WithFields(log.Fields{
					"idx":   idx,
					"batch": batch,
					"type":  "failed",
					"error": err,
				}).Error("inserting batch")
				totalFailed += len(batch)
			} else {
				log.WithFields(log.Fields{
					"batch_size": batchSize,
					"type":       "success",
				}).Info("inserting batch")
				totalSuccess += len(batch)
			}
			batch = batch[:0]
		}
	}

	if len(batch) > 0 {
		err := executeBatch()
		if err != nil {
			log.WithFields(log.Fields{
				"batch": batch,
				"type":  "failed",
				"error": err,
			}).Error("inserting batch")
			totalFailed += len(batch)
		} else {
			log.WithFields(log.Fields{
				"batch_size": len(batch),
				"type":       "success",
			}).Info("inserting batch")
			totalSuccess += len(batch)
		}
	}

	return totalSuccess, totalFailed, nil
}

func init() {
	plugins.RegisterDestination("timescaledb", func() plugins.Destination {
		return &TimescaleDBDestination{
			BatchSize: 1000,
		}
	})
}
