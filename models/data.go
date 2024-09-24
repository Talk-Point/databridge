package models

import (
	"errors"
	"fmt"
	"log"
)

type ColumnType int

const (
	String ColumnType = iota
	BigInt
	Float
	DateTime
	Int
)

func (ct ColumnType) String() string {
	return [...]string{"string", "bigint", "float", "datetime", "int"}[ct]
}

func ParseColumnType(s string) (ColumnType, error) {
	switch s {
	case "string":
		return String, nil
	case "bigint":
		return BigInt, nil
	case "float":
		return Float, nil
	case "datetime":
		return DateTime, nil
	case "int":
		return Int, nil
	default:
		return -1, fmt.Errorf("invalid column type: %s", s)
	}
}

type Column struct {
	Name string
	Type ColumnType
}

type Model struct {
	Columns []Column
	Unique  []string
}

func LoadModel(data map[string]interface{}) (*Model, error) {
	model := &Model{}

	columns, ok := data["columns"].([]interface{})
	if !ok {
		return nil, errors.New("invalid columns format")
	}

	for _, column := range columns {
		columnData, ok := column.(map[interface{}]interface{})
		if !ok {
			log.Printf("Invalid column data: %v\n", column)
			return nil, errors.New("invalid column data format")
		}

		name, ok := columnData["name"].(string)
		if !ok {
			return nil, errors.New("invalid column name format")
		}

		typeStr, ok := columnData["type"].(string)
		if !ok {
			return nil, errors.New("invalid column type format")
		}

		columnType, err := ParseColumnType(typeStr)
		if err != nil {
			return nil, err
		}

		model.Columns = append(model.Columns, Column{
			Name: name,
			Type: columnType,
		})
	}

	unique, ok := data["unique_key"].([]interface{})
	if !ok {
		return nil, errors.New("invalid unique_key format")
	}

	for _, key := range unique {
		keyStr, ok := key.(string)
		if !ok {
			return nil, errors.New("invalid unique_key entry format")
		}
		model.Unique = append(model.Unique, keyStr)
	}

	return model, nil
}
