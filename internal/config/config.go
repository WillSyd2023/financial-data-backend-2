package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config is the top-level struct that holds all configuration.
type Config struct {
	APIPort  string        `yaml:"api_port"`
	Finnhub  FinnhubConfig `yaml:"finnhub"`
	Kafka    KafkaConfig   `yaml:"kafka"`
	MongoDB  MongoConfig   `yaml:"mongodb"`
	Symbols  []string      `yaml:"subscribed_symbols"`
	Timeouts TimeoutConfig `yaml:"timeouts"`
}

// FinnhubConfig holds the configuration for the Finnhub API.
type FinnhubConfig struct {
	Token string `yaml:"token"`
}

// KafkaConfig holds the configuration for the Kafka connection.
type KafkaConfig struct {
	BrokerURL string `yaml:"broker_url"`
	Topic     string `yaml:"topic"`
}

// MongoConfig holds the configuration for the MongoDB cloud storage.
type MongoConfig struct {
	URL                   string `yaml:"url"`
	DatabaseName          string `yaml:"database_name"`
	CollectionName        string `yaml:"collection_name"`
	SymbolsCollectionName string `yaml:"symbols_collection_name"`
}

type TimeoutConfig struct {
	APIRequest          time.Duration `yaml:"api_request"`
	BackgroundOperation time.Duration `yaml:"background_operation"`
	Shutdown            time.Duration `yaml:"shutdown"`
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
