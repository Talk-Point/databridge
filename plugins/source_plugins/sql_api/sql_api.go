package sql_api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Talk-Point/databridge/plugins"
)

type SQLAPISource struct {
	Endpoint string
	APIToken string
	Query    string
	Date     string
}

func (s *SQLAPISource) Init(config map[string]interface{}) error {
	// Extract and validate configuration parameters
	s.Endpoint = config["endpoint"].(string)
	s.APIToken = os.Getenv("API_TOKEN")
	if s.APIToken == "" {
		return fmt.Errorf("API_TOKEN environment variable is required")
	}
	s.Query = config["query"].(string)
	// Handle dynamic date parameter
	s.Date = time.Now().Format("02.01.2006")
	return nil
}

func (s *SQLAPISource) FetchData() ([]map[string]interface{}, error) {
	// Replace {date} in the query
	query := strings.ReplaceAll(s.Query, "{date}", s.Date)

	// Prepare the request
	reqBody := map[string]string{"query": query}
	reqJSON, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", s.Endpoint, bytes.NewBuffer(reqJSON))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-KEY", s.APIToken)

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to fetch data: %s", string(bodyBytes))
	}

	// Parse response
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	// Extract data
	data, ok := result["results"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}

	// Convert to []map[string]interface{}
	records := make([]map[string]interface{}, 0, len(data))
	for _, item := range data {
		record, ok := item.(map[string]interface{})
		if !ok {
			// @todo logging error convertion
			continue
		}
		records = append(records, record)
	}

	return records, nil
}

func init() {
	plugins.RegisterSource("sql_api", func() plugins.Source {
		return &SQLAPISource{}
	})
}
