package timescaledb

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/Talk-Point/databridge/models"
	"github.com/Talk-Point/databridge/plugins"
	_ "github.com/lib/pq"
)

type TimescaleDBDestination struct {
	DB     *sql.DB
	Table  string
	Schema map[string]string // Column types
}

func (d *TimescaleDBDestination) Init(config map[string]interface{}, model *models.Model) error {
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

func (d *TimescaleDBDestination) StoreData(data []map[string]interface{}) error {
	fmt.Println("Storing data to TimescaleDB")
	return nil
}

func init() {
	plugins.RegisterDestination("timescaledb", func() plugins.Destination {
		return &TimescaleDBDestination{}
	})
}
