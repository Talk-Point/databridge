package pkg

import (
	"flag"
	"log"
	"time"
)

type TimePartitionParams struct {
	LogLevel   string
	ConfigPath string
	RunSchema  bool
	Date       string
	Start      string
	End        string
	Interval   string
	Location   *time.Location
	StartTime  time.Time
	EndTime    time.Time
	Kestra     bool
	FilePath   string
}

func NewTimePartitionParams() *TimePartitionParams {
	return &TimePartitionParams{
		Location:  time.Local,
		Kestra:    false,
		LogLevel:  "info",
		RunSchema: false,
	}
}

func (p *TimePartitionParams) ParseFlags() error {
	// Define CLI parameters
	flag.StringVar(&p.LogLevel, "log-level", "info", "Log level (debug, info, warn, error, fatal, panic)")
	flag.StringVar(&p.ConfigPath, "config", "config.yaml", "Path to configuration file")
	flag.BoolVar(&p.RunSchema, "run-schema", false, "run schema query")
	flag.StringVar(&p.Date, "date", "", "Date (e.g., 2024-09-01)")
	flag.StringVar(&p.Start, "start", "", "Start time in RFC3339 format (e.g., 2024-09-01T00:00:00Z)")
	flag.StringVar(&p.End, "end", "", "End time in RFC3339 format (e.g., 2024-09-02T00:00:00Z)")
	flag.StringVar(&p.Interval, "interval", "", "Interval duration (e.g., 30m for 30 minutes)")
	flag.StringVar(&p.FilePath, "file-path", "", "Path to file")
	flag.BoolVar(&p.Kestra, "kestra", false, "Output kestra metrics")

	// Parse CLI flags
	flag.Parse()

	var err error

	// parse date if provided
	if p.Date != "" {
		p.StartTime, err = time.Parse("2006-01-02", p.Date)
		if err != nil {
			log.Fatalf("Invalid date format: %v", err)
			return err
		}
		p.EndTime = p.StartTime.AddDate(0, 0, 1)
	}

	// Parse start and end times if provided
	if p.Start != "" {
		p.StartTime, err = time.Parse(time.RFC3339, p.Start)
		if err != nil {
			log.Fatalf("Invalid start time format: %v", err)
			return err
		}
	}

	if p.End != "" {
		p.EndTime, err = time.Parse(time.RFC3339, p.End)
		if err != nil {
			log.Fatalf("Invalid end time format: %v", err)
			return err
		}
	}

	// If interval is provided, calculate start time based on the interval and the current time or end time
	if p.Interval != "" {
		intervalDuration, err := time.ParseDuration(p.Interval)
		if err != nil {
			log.Fatalf("Invalid interval format: %v", err)
			return err
		}

		p.EndTime = time.Now()
		p.StartTime = p.EndTime.Add(-intervalDuration)
	}

	return nil
}
