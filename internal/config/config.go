package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config is the top-level struct that holds all configuration.
type Config struct {
	Finnhub FinnhubConfig `yaml:"finnhub"` // Changed from Alpaca
	Kafka   KafkaConfig   `yaml:"kafka"`
	Symbols []string      `yaml:"subscribed_symbols"`
}

// FinnhubConfig holds the configuration for the Finnhub API.
type FinnhubConfig struct {
	Token string `yaml:"token"` // Changed from Key/Secret
}

// KafkaConfig holds the configuration for the Kafka connection.
type KafkaConfig struct {
	BrokerURL string `yaml:"broker_url"`
	Topic     string `yaml:"topic"`
}

// LoadConfig reads the configuration file from the given path and
// returns a Config struct. (This function does not need to change).
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
