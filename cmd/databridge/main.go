package main

import (
	"os"

	"github.com/Talk-Point/databridge/models"
	"github.com/Talk-Point/databridge/pkg"
	"github.com/Talk-Point/databridge/pkg/kestra"
	_ "github.com/Talk-Point/databridge/plugins/destination_plugins/timescaledb"
	_ "github.com/Talk-Point/databridge/plugins/source_plugins/csv_v1"
	_ "github.com/Talk-Point/databridge/plugins/source_plugins/sql_api"

	"github.com/Talk-Point/databridge/config"
	"github.com/Talk-Point/databridge/plugins"
	log "github.com/sirupsen/logrus"
)

func run(flags *pkg.TimePartitionParams) {
	// Load configuration
	cfg, err := config.LoadConfig(flags.ConfigPath)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
		os.Exit(1)
	}

	// initialize model
	model, err := models.LoadModel(cfg.Model.Config)
	if err != nil {
		log.Fatalf("Error loading model: %v", err)
		os.Exit(1)
	}

	// Initialize source plugin
	source, err := plugins.GetSource(cfg.Source.Type)
	if err != nil {
		log.Fatalf("Error getting source plugin: %v", err)
		os.Exit(1)
	}
	err = source.Init(cfg.Source.Config, model)
	if err != nil {
		log.Fatalf("Error initializing source plugin: %v", err)
		os.Exit(1)
	}

	// Initialize destination plugin
	destination, err := plugins.GetDestination(cfg.Destination.Type)
	if err != nil {
		log.Fatalf("Error getting destination plugin: %v", err)
		os.Exit(1)
	}
	err = destination.Init(cfg.Destination.Config, model)
	if err != nil {
		log.Fatalf("Error initializing destination plugin: %v", err)
		os.Exit(1)
	}

	if flags.RunSchema {
		log.Info("destination schema query requested")
		err = destination.RunSchema()
		if err != nil {
			log.Fatalf("Error running schema query: %v", err)
			os.Exit(1)
		}
	}

	// Fetch data
	data, err := source.FetchData(map[string]interface{}{
		"start_at":  flags.StartTime,
		"end_at":    flags.EndTime,
		"file_path": flags.FilePath,
	})
	if err != nil {
		log.Fatalf("Error fetching data: %v", err)
		os.Exit(1)
	}

	// Store data
	totalSuccess, totalErrored, err := destination.StoreData(data)
	if err != nil {
		log.Fatalf("Error storing data: %v", err)
		os.Exit(1)
	}

	if flags.Kestra {
		kestra.CounterMetric("total", float64(totalSuccess)).
			WithTags(map[string]string{"status": "success"}).
			Log()
		kestra.CounterMetric("total", float64(totalErrored)).
			WithTags(map[string]string{"status": "errored"}).
			Log()
	}
	if totalErrored > 0 {
		log.WithFields(log.Fields{
			"total_success": totalSuccess,
			"total_errored": totalErrored,
		}).Error("data transfer completed with errors.")
		os.Exit(1)
	} else {
		log.WithFields(log.Fields{
			"total_success": totalSuccess,
			"total_errored": totalErrored,
		}).Info("data transfer completed successfully.")
	}
}

func main() {
	// Parse Flags
	flags := pkg.NewTimePartitionParams()
	flags.ParseFlags()

	if flags.LogLevel != "" {
		level, err := log.ParseLevel(flags.LogLevel)
		if err != nil {
			log.Fatalf("Invalid log level: %v", err)
			os.Exit(1)
		}
		log.SetLevel(level)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	run(flags)
}
