package plugins

import (
	"fmt"

	"github.com/Talk-Point/databridge/models"
	log "github.com/sirupsen/logrus"
)

// Source interface
type Source interface {
	Init(config map[string]interface{}, model *models.Model) error
	FetchData(opts map[string]interface{}) ([]map[string]interface{}, error)
	Close() error
}

// Destination interface
type Destination interface {
	Init(config map[string]interface{}, model *models.Model) error
	StoreData(data []map[string]interface{}) (int, int, error)
	RunSchema() error
	Close() error
}

type SourceFactory func() Source
type DestinationFactory func() Destination

var (
	sourceFactories      = make(map[string]SourceFactory)
	destinationFactories = make(map[string]DestinationFactory)
)

func RegisterSource(name string, factory SourceFactory) {
	log.WithFields(log.Fields{
		"name": name,
	}).Debug("Registering source plugin: ", name)
	sourceFactories[name] = factory
}

func GetSource(name string) (Source, error) {
	log.WithFields(log.Fields{
		"sources": sourceFactories,
	}).Debug("Registry possible sources")
	factory, ok := sourceFactories[name]
	if !ok {
		return nil, fmt.Errorf("source plugin '%s' not found", name)
	}
	return factory(), nil
}

func RegisterDestination(name string, factory DestinationFactory) {
	log.WithFields(log.Fields{
		"name": name,
	}).Debug("Registering destination plugin: ", name)
	destinationFactories[name] = factory
}

func GetDestination(name string) (Destination, error) {
	log.WithFields(log.Fields{
		"destinations": destinationFactories,
	}).Debug("Registry possible destinations")
	factory, ok := destinationFactories[name]
	if !ok {
		return nil, fmt.Errorf("destination plugin '%s' not found", name)
	}
	return factory(), nil
}
