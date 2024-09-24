package sql_api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Talk-Point/databridge/models"
	"github.com/Talk-Point/databridge/plugins"
	log "github.com/sirupsen/logrus"
)

type SQLAPISource struct {
	Model    *models.Model
	Endpoint string
	APIToken string
	Query    string
	Date     string
}

func (s *SQLAPISource) Init(config map[string]interface{}, model *models.Model) error {
	s.Model = model
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
	total_errored := 0
	records := make([]map[string]interface{}, 0, len(data))
	for _, item := range data {
		transformedRecord, err := s.Transform(item)
		if err != nil {
			log.WithFields(log.Fields{
				"item": item,
				"err":  err,
			}).Errorf("Error transforming record: %v", err)
			total_errored++
			continue
		}
		records = append(records, transformedRecord)
	}

	return records, nil
}

func (s *SQLAPISource) Close() error {
	return nil
}

func convert(value string, columnType models.ColumnType) (interface{}, error) {
	switch columnType {
	case models.String:
		return value, nil
	case models.BigInt:
		// convert string to bigint
		value, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, err
		}
		return value, nil
	case models.Float:
		// convert string to float
		// replace "," with "." to parse float
		value, err := strconv.ParseFloat(strings.Replace(value, ",", ".", -1), 64)
		if err != nil {
			return nil, err
		}
		return value, nil
	case models.DateTime:
		// convert string to datetime
		loc, err := time.LoadLocation("Europe/Berlin")
		if err != nil {
			return nil, err
		}
		value, err := time.ParseInLocation("02.01.2006 15:04:05", value, loc)
		if err != nil {
			return nil, err
		}
		return value, nil
	case models.Int:
		// convert string to int
		value, err := strconv.Atoi(value)
		if err != nil {
			return nil, err
		}
		return value, nil
	default:
		return nil, fmt.Errorf("invalid column type: %s", columnType)
	}
}

func (s *SQLAPISource) Transform(item interface{}) (map[string]interface{}, error) {
	record, ok := item.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected record format")
	}

	for _, column := range s.Model.Columns {
		if _, ok := record[column.Name]; ok {
			data, err := convert(record[column.Name].(string), column.Type)
			if err != nil {
				continue
			}
			record[column.Name] = data
		} else {
			log.WithField("column", column.Name).Warn("Column not found in record")
			record[column.Name] = nil
		}
	}

	return record, nil
}

func init() {
	plugins.RegisterSource("sql_api", func() plugins.Source {
		return &SQLAPISource{}
	})
}
