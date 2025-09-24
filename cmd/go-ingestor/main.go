package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/url"
	"strconv"
	"time"

	"financial-data-backend-2/internal/config"

	"github.com/gorilla/websocket"
	"github.com/segmentio/kafka-go"
)

func ensureTopic(cfg config.KafkaConfig) error {
	// Dial the Kafka broker to create a connection for administrative tasks
	conn, err := kafka.Dial("tcp", cfg.BrokerURL)
	if err != nil {
		log.Printf("Failed to dial Kafka for topic creation: %v", err)
		return err
	}
	defer conn.Close()

	// Get the controller broker
	controller, err := conn.Controller()
	if err != nil {
		log.Printf("Failed to get Kafka controller: %v", err)
		return err
	}

	// Connect to the controller broker
	controllerConn, err := kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		log.Printf("Failed to connect to Kafka controller: %v", err)
		return err
	}
	defer controllerConn.Close()

	// Define the topic configuration
	topicConfig := kafka.TopicConfig{
		Topic:             cfg.Topic,
		NumPartitions:     1,
		ReplicationFactor: 1,
	}

	// Create the topic
	err = controllerConn.CreateTopics(topicConfig)
	if err != nil {
		log.Printf("Failed to create Kafka topic: %v", err)
		return err
	}

	log.Printf("Kafka topic '%s' is ready", cfg.Topic)
	return nil
}

func main() {
	// - Load Configuration
	cfg, err := config.LoadConfig("config/config.yml")
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// - Retry loop to wait for Kafka to be truly ready.
	for {
		err := ensureTopic(cfg.Kafka)
		if err == nil {
			break
		}
		log.Println("Could not ensure Kafka topic exists, retrying in 2 seconds...")
		time.Sleep(2 * time.Second)
	}

	// - Establish WebSocket Connection
	u := url.URL{Scheme: "wss", Host: "ws.finnhub.io", RawQuery: "token=" + cfg.Finnhub.Token}
	// log.Printf("Connecting to %s", u.String())

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer conn.Close()
	log.Println("Successfully connected to Finnhub WebSocket")

	// FinnHub has given ping messages before
	conn.SetPingHandler(nil)

	// - Subscribe to Symbols
	// Iterate over the symbols from your config file
	for _, symbol := range cfg.Symbols {
		subMsg, _ := json.Marshal(map[string]interface{}{"type": "subscribe", "symbol": symbol})
		log.Printf("Subscribing to %s", symbol)
		if err := conn.WriteMessage(websocket.TextMessage, subMsg); err != nil {
			log.Fatalf("Failed to subscribe to %s: %v", symbol, err)
		}
	}

	// 4. Setup Kafka
	kafkaWriter := &kafka.Writer{
		Addr:     kafka.TCP(cfg.Kafka.BrokerURL),
		Topic:    cfg.Kafka.Topic,
		Balancer: &kafka.LeastBytes{},
	}
	defer kafkaWriter.Close()
	log.Println("Kafka writer configured successfully")

	// - The Kafka Read Loop
	log.Println("Waiting for messages...")
	for {
		// Read a message from the connection
		var message interface{}
		err := conn.ReadJSON(&message)
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		// Skip logging pings
		msgMap, ok := message.(map[string]interface{})
		if ok && msgMap["type"] == "ping" {
			continue
		}

		// Otherwise, send message to Kafka
		msgBytes, err := json.Marshal(message)
		if err != nil {
			log.Printf("Failed to marshal message: %v", err)
			continue
		}
		err = kafkaWriter.WriteMessages(context.Background(),
			kafka.Message{Value: msgBytes})
		if err != nil {
			log.Printf("Failed to write message to Kafka: %v", err)
		} else {
			log.Printf("Successfully sent message to Kafka: %s", string(msgBytes))
		}
	}
}
