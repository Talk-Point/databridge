package main

import (
	"flag"
	"log"
	"os"

	_ "github.com/Talk-Point/databridge/plugins/destination_plugins/timescaledb"
	_ "github.com/Talk-Point/databridge/plugins/source_plugins/sql_api"

	"github.com/Talk-Point/databridge/config"
	"github.com/Talk-Point/databridge/plugins"
)

func main() {
	// Parse flags to get the config file path
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
		os.Exit(1)
	}

	// Initialize source plugin
	source, err := plugins.GetSource(cfg.Source.Type)
	if err != nil {
		log.Fatalf("Error getting source plugin: %v", err)
		os.Exit(1)
	}
	err = source.Init(cfg.Source.Config)
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
	err = destination.Init(cfg.Destination.Config)
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
	err = destination.StoreData(data)
	if err != nil {
		log.Fatalf("Error storing data: %v", err)
		os.Exit(1)
	}

	log.Println("Data transfer completed successfully.")
}
