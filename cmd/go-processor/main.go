package main

import (
	"context"
	"encoding/json"
	"financial-data-backend-2/internal/config"
	"financial-data-backend-2/internal/kafka"
	"financial-data-backend-2/internal/models"
	"log"
	"time"

	kafkaGo "github.com/segmentio/kafka-go"
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

	// - Setup Kafka Reader
	// Create the Kafka Reader
	r := kafkaGo.NewReader(kafkaGo.ReaderConfig{
		Brokers: []string{cfg.Kafka.BrokerURL},
		Topic:   cfg.Kafka.Topic,
		GroupID: "finnhub-websocket-consumer-group",
		//    Essential for distributed consumption and offset tracking
		// MaxBytes:    10e6,
		//    Optional: Maximum amount of bytes to fetch in a single request (10MB)
		// StartOffset: kafkaGo.FirstOffset,
		//    Optional: Start from the beginning if no committed offset is found for the GroupID
	})
	defer func() {
		if err := r.Close(); err != nil {
			log.Fatal("failed to close Kafka Reader:", err)
		}
		log.Println("Kafka Reader closed.")
	}()
	log.Println(`Kafka reader configured successfully. 
	Consumer Group ID: finnhub-websocket-consumer-group`)

	// - The Kafka Read Loop
	log.Println("Waiting for messages...")
	for {
		// Read a message from Kafka
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		log.Printf("Message received | Topic: %s | Partition: %d | Offset: %d\n",
			m.Topic, m.Partition, m.Offset)
		log.Printf("Message Value: %s", string(m.Value))

		// Unmarshal the raw JSON value from Kafka
		var finnMsg models.FinnhubTradeMessage
		if err := json.Unmarshal(m.Value, &finnMsg); err != nil {
			log.Printf("Failed to unmarshal message from Kafka: %v", err)
			log.Printf("Bad Message Value: %s", string(m.Value))
			continue
		}

		// Check the message type
		if finnMsg.Type != "trade" {
			log.Printf("Received non-trade message type: %s. Skipping.",
				finnMsg.Type)
			continue
		}
	}
}
