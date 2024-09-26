package kestra

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
)

// TestCounterMetricCreation validates that the CounterMetric function correctly creates a metric.
func TestCounterMetricCreation(t *testing.T) {
	metric := CounterMetric("total_success", 5)

	if metric.Name != "total_success" {
		t.Errorf("expected name to be 'total_success', got %s", metric.Name)
	}

	if metric.Type != Counter {
		t.Errorf("expected type to be 'counter', got %s", metric.Type)
	}

	if metric.Value != 5 {
		t.Errorf("expected value to be 5, got %f", metric.Value)
	}

	if len(metric.Tags) != 0 {
		t.Errorf("expected no tags, got %d", len(metric.Tags))
	}
}

// TestMetricWithTags validates that tags are properly added to the metric.
func TestMetricWithTags(t *testing.T) {
	metric := CounterMetric("total_success", 5)
	tags := map[string]string{
		"env":  "production",
		"host": "server-1",
	}

	metric.WithTags(tags)

	if len(metric.Tags) != len(tags) {
		t.Errorf("expected %d tags, got %d", len(tags), len(metric.Tags))
	}

	for key, value := range tags {
		if metric.Tags[key] != value {
			t.Errorf("expected tag %s to be %s, got %s", key, value, metric.Tags[key])
		}
	}
}

// TestMetricLog validates that the Log function outputs the correct format for Kestra.
func TestMetricLog(t *testing.T) {
	metric := CounterMetric("total_success", 5)
	tags := map[string]string{
		"env":  "production",
		"host": "server-1",
	}
	metric.WithTags(tags)

	// Capture output
	output := captureOutput(func() {
		metric.Log()
	})

	// Validate output format
	expectedPrefix := "::"
	expectedSuffix := "::\n"
	if !strings.HasPrefix(output, expectedPrefix) || !strings.HasSuffix(output, expectedSuffix) {
		t.Errorf("output should be wrapped with '::' and end with newline, got: %s", output)
	}

	// Validate JSON structure
	trimmedOutput := strings.TrimPrefix(output, expectedPrefix)
	trimmedOutput = strings.TrimSuffix(trimmedOutput, expectedSuffix)

	var kestraMetric KestraMetric
	err := json.Unmarshal([]byte(trimmedOutput), &kestraMetric)
	if err != nil {
		t.Fatalf("failed to unmarshal output into KestraMetric: %v", err)
	}

	if len(kestraMetric.Metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(kestraMetric.Metrics))
	}

	if kestraMetric.Metrics[0].Name != "total_success" {
		t.Errorf("expected name to be 'total_success', got %s", kestraMetric.Metrics[0].Name)
	}

	if kestraMetric.Metrics[0].Type != Counter {
		t.Errorf("expected type to be 'counter', got %s", kestraMetric.Metrics[0].Type)
	}

	if kestraMetric.Metrics[0].Value != 5 {
		t.Errorf("expected value to be 5, got %f", kestraMetric.Metrics[0].Value)
	}

	for key, value := range tags {
		if kestraMetric.Metrics[0].Tags[key] != value {
			t.Errorf("expected tag %s to be %s, got %s", key, value, kestraMetric.Metrics[0].Tags[key])
		}
	}
}

// Helper function to capture the output of the Log function.
func captureOutput(f func()) string {
	// Create a pipe to capture standard output
	r, w, _ := os.Pipe()
	originalStdout := os.Stdout
	os.Stdout = w

	// Execute the function while redirecting the output
	f()

	// Restore the original stdout and close the pipe writer
	w.Close()
	os.Stdout = originalStdout

	// Read the captured output from the pipe reader
	var buf bytes.Buffer
	io.Copy(&buf, r)

	return buf.String()
}
