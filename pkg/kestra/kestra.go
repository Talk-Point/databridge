package kestra

import (
	"encoding/json"
	"fmt"
	"log"
)

type MetricType string

const (
	Counter MetricType = "counter"
	Timer   MetricType = "timer"
)

type Metric struct {
	Name  string            `json:"name"`
	Type  MetricType        `json:"type"`
	Value float64           `json:"value"`
	Tags  map[string]string `json:"tags,omitempty"`
}

// KestraMetric is a struct to hold the metric details.
type KestraMetric struct {
	Metrics []Metric `json:"metrics"`
}

// KestraCounterMetric creates a new counter metric with the provided name and value.
//
// Example usage:
//
//	metric := CounterMetric("total_success", 5)
//	metric.Log()
//
// Params:
//   - name: The name of the counter metric (e.g., "total_success").
//   - value: The value for the counter metric.
//
// Returns:
//   - A pointer to the newly created Metric struct.
func CounterMetric(name string, value float64) *Metric {
	return &Metric{
		Name:  name,
		Type:  Counter,
		Value: value,
		Tags:  make(map[string]string),
	}
}

// WithTags allows you to add tags to the metric.
func (m *Metric) WithTags(tags map[string]string) *Metric {
	for key, value := range tags {
		m.Tags[key] = value
	}
	return m
}

// Log prints the metric in a format that Kestra can capture.
func (m *Metric) Log() {
	kestraMetric := KestraMetric{
		Metrics: []Metric{*m},
	}

	// Convert struct to JSON
	jsonOutput, err := json.Marshal(kestraMetric)
	if err != nil {
		log.Fatalf("Failed to marshal metric: %v", err)
	} else {
		fmt.Printf("::%s::\n", string(jsonOutput))
	}
}
