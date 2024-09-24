package plugins

import (
	"fmt"

	"github.com/Talk-Point/databridge/models"
)

// Source interface
type Source interface {
	Init(config map[string]interface{}, model *models.Model) error
	FetchData() ([]map[string]interface{}, error)
	Close() error
}

// Destination interface
type Destination interface {
	Init(config map[string]interface{}, model *models.Model) error
	StoreData(data []map[string]interface{}) (int, int, error)
	Close() error
}

type SourceFactory func() Source
type DestinationFactory func() Destination

var (
	sourceFactories      = make(map[string]SourceFactory)
	destinationFactories = make(map[string]DestinationFactory)
)

func RegisterSource(name string, factory SourceFactory) {
	sourceFactories[name] = factory
}

func GetSource(name string) (Source, error) {
	factory, ok := sourceFactories[name]
	if !ok {
		return nil, fmt.Errorf("source plugin '%s' not found", name)
	}
	return factory(), nil
}

func RegisterDestination(name string, factory DestinationFactory) {
	destinationFactories[name] = factory
}

func GetDestination(name string) (Destination, error) {
	factory, ok := destinationFactories[name]
	if !ok {
		return nil, fmt.Errorf("destination plugin '%s' not found", name)
	}
	return factory(), nil
}
