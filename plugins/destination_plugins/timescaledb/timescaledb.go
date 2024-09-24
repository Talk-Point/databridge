package timescaledb

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/Talk-Point/databridge/models"
	"github.com/Talk-Point/databridge/plugins"
	_ "github.com/lib/pq"
)

type TimescaleDBDestination struct {
	Model  *models.Model
	DB     *sql.DB
	Table  string
	Schema map[string]string // Column types
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

func (d *TimescaleDBDestination) GetSchema() ([]string, error) {
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
		stm.WriteString(", time)\n")
	}

	stm.WriteString(");\n")

	create_hypertable := fmt.Sprintf("SELECT create_hypertable('%s', 'time', if_not_exists => TRUE);", d.Table)

	return []string{stm.String(), create_hypertable}, nil
}

func (d *TimescaleDBDestination) StoreData(data []map[string]interface{}) error {
	fmt.Println("Storing data to TimescaleDB")
	return nil
}

func init() {
	plugins.RegisterDestination("timescaledb", func() plugins.Destination {
		return &TimescaleDBDestination{}
	})
}
