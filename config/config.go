package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

// Define structs matching your configuration schema
type Config struct {
	Name        string       `yaml:"name"`
	Source      PluginConfig `yaml:"source"`
	Destination PluginConfig `yaml:"destination"`
	Model       PluginConfig `yaml:"model"`
}

type PluginConfig struct {
	Type   string                 `yaml:"type"`
	Config map[string]interface{} `yaml:",inline"` // Capture other keys
}

// Load configuration
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
