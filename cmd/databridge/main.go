package main

import (
	"flag"
	"os"

	"github.com/Talk-Point/databridge/models"
	_ "github.com/Talk-Point/databridge/plugins/destination_plugins/timescaledb"
	_ "github.com/Talk-Point/databridge/plugins/source_plugins/sql_api"

	"github.com/Talk-Point/databridge/config"
	"github.com/Talk-Point/databridge/plugins"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetLevel(log.DebugLevel)

	// Parse flags to get the config file path
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
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

	// Fetch data
	data, err := source.FetchData()
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

	if totalErrored > 0 {
		log.WithFields(log.Fields{
			"total_success": totalSuccess,
			"total_errored": totalErrored,
		}).Error("Data transfer completed with errors.")
		os.Exit(1)
	} else {
		log.WithFields(log.Fields{
			"total_success": totalSuccess,
			"total_errored": totalErrored,
		}).Info("Data transfer completed successfully.")
	}
}
