package csv

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Talk-Point/databridge/models"
	"github.com/Talk-Point/databridge/plugins"
	log "github.com/sirupsen/logrus"
)

type CSVSource struct {
	Model *models.Model
}

func (s *CSVSource) Init(config map[string]interface{}, model *models.Model) error {
	s.Model = model
	return nil
}

func (s *CSVSource) FetchData(opts map[string]interface{}) ([]map[string]interface{}, error) {
	filepath, ok := opts["file_path"].(string)
	if !ok {
		return nil, errors.New("file_path is missing")
	}

	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1 // Allow variable number of fields per record

	// Read the header
	header, err := reader.Read()
	if err != nil {
		return nil, err
	}

	// Read the data
	var records []map[string]interface{}
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.WithError(err).Error("Error reading CSV row")
			continue
		}

		// Map row to record
		record := make(map[string]interface{})
		for i, value := range row {
			if i < len(header) {
				columnName := header[i]
				record[columnName] = value
			}
		}

		transformedRecord, err := s.Transform(record)
		if err != nil {
			log.WithFields(log.Fields{
				"record": record,
				"error":  err,
			}).Error("Error transforming record")
			continue
		}

		records = append(records, transformedRecord)
	}

	return records, nil
}

func (s *CSVSource) Transform(record map[string]interface{}) (map[string]interface{}, error) {
	for _, column := range s.Model.Columns {
		if val, ok := record[column.Name]; ok {
			strVal, ok := val.(string)
			if !ok {
				return nil, fmt.Errorf("value for column %s is not a string", column.Name)
			}
			data, err := convert(strVal, column.Type)
			if err != nil {
				return nil, fmt.Errorf("error converting column %s: %v", column.Name, err)
			}
			record[column.Name] = data
		} else {
			log.WithField("column", column.Name).Warn("Column not found in record")
			record[column.Name] = nil
		}
	}
	return record, nil
}

func convert(value string, columnType models.ColumnType) (interface{}, error) {
	switch columnType {
	case models.String:
		return value, nil
	case models.BigInt:
		intVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, err
		}
		return intVal, nil
	case models.Float:
		floatVal, err := strconv.ParseFloat(strings.Replace(value, ",", ".", -1), 64)
		if err != nil {
			return nil, err
		}
		return floatVal, nil
	case models.DateTime:
		loc, err := time.LoadLocation("Europe/Berlin")
		if err != nil {
			return nil, err
		}
		dateVal, err := time.ParseInLocation("2006-01-02T15:04:05", value, loc)
		if err != nil {
			return nil, err
		}
		return dateVal, nil
	case models.DateTimeNullable:
		if value == "" {
			return nil, nil
		}
		loc, err := time.LoadLocation("Europe/Berlin")
		if err != nil {
			return nil, err
		}
		dateVal, err := time.ParseInLocation("2006-01-02T15:04:05", value, loc)
		if err != nil {
			return nil, err
		}
		return dateVal, nil
	case models.Int:
		intVal, err := strconv.Atoi(value)
		if err != nil {
			return nil, err
		}
		return intVal, nil
	default:
		return nil, fmt.Errorf("invalid column type: %s", columnType)
	}
}

func (s *CSVSource) Close() error {
	return nil
}

func init() {
	plugins.RegisterSource("csv", func() plugins.Source {
		return &CSVSource{}
	})
}
