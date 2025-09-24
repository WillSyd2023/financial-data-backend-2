package kafka

import (
	"financial-data-backend-2/internal/config"
	"log"
	"net"
	"strconv"

	kafkaGo "github.com/segmentio/kafka-go"
)

func EnsureTopic(cfg config.KafkaConfig) error {
	// Dial the Kafka broker to create a connection for administrative tasks
	conn, err := kafkaGo.Dial("tcp", cfg.BrokerURL)
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
	controllerConn, err := kafkaGo.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		log.Printf("Failed to connect to Kafka controller: %v", err)
		return err
	}
	defer controllerConn.Close()

	// Define the topic configuration
	topicConfig := kafkaGo.TopicConfig{
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
