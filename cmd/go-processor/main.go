package main

import (
	"financial-data-backend-2/internal/config"
	"financial-data-backend-2/internal/kafka"
	"log"
	"time"
)

func main() {
	// - Load Configuration
	cfg, err := config.LoadConfig("config/config.yml")
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// - Retry loop to wait for Kafka to be truly ready.
	for {
		err := kafka.EnsureTopic(cfg.Kafka)
		if err == nil {
			break
		}
		log.Println("Could not ensure Kafka topic exists, retrying in 2 seconds...")
		time.Sleep(2 * time.Second)
	}
}
