package main

import (
	"context"
	"encoding/json"
	"financial-data-backend-2/internal/config"
	"financial-data-backend-2/internal/kafka"
	"financial-data-backend-2/internal/models"
	mongoGo "financial-data-backend-2/internal/mongo"
	"log"
	"strconv"
	"time"

	kafkaGo "github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

	// - Setup MongoDB database
	DB, err := mongoGo.ConnectDB(cfg.MongoDB.URL)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := DB.Disconnect(ctx); err != nil {
			log.Fatalf("Error during MongoDB disconnect: %v", err)
		}
		log.Println("MongoDB client disconnected.")
	}()

	// - Setup collections
	symbolCollection := mongoGo.GetCollection(DB, cfg.MongoDB.DatabaseName,
		cfg.MongoDB.SymbolsCollectionName)
	tradeCollection := mongoGo.GetCollection(DB, cfg.MongoDB.DatabaseName,
		cfg.MongoDB.CollectionName)

	// Ensure a unique index exists on the symbol collection for performance.
	// This is a one-time setup.
	_, err = symbolCollection.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys:    bson.M{"symbol": 1},
			Options: options.Index().SetUnique(true),
		},
	)
	if err != nil {
		log.Printf("Could not create unique index on symbols (may already exist): %v", err)
	}

	// - The Read Loop
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

		// Check the message
		if finnMsg.Type != "trade" {
			log.Printf("Received non-trade message type: %s. Skipping.",
				finnMsg.Type)
			continue
		}
		if len(finnMsg.Data) == 0 {
			continue
		}

		// Prepare data for insertion/updates
		timeSeries := make([]any, len(finnMsg.Data))
		symbolTradeCounts := make(map[string]int64)
		latestTimestamps := make(map[string]time.Time)
		for i, trade := range finnMsg.Data {
			// time
			t := time.UnixMilli(trade.Timestamp)

			// price
			pStr := strconv.FormatFloat(trade.Price, 'f', -1, 64)
			p, err := primitive.ParseDecimal128(pStr)
			if err != nil {
				log.Printf("Could not convert price string '%s' for symbol '%s' to Decimal128: %v",
					pStr, trade.Symbol, err)
				continue // Skip this tick if the price is invalid
			}

			// volume
			vStr := strconv.FormatFloat(trade.Volume, 'f', -1, 64)
			v, err := primitive.ParseDecimal128(vStr)
			if err != nil {
				log.Printf("Could not convert volume string '%s' for symbol '%s' to Decimal128: %v",
					vStr, trade.Symbol, err)
				continue // Skip this tick if the volume is invalid
			}

			// put trade to batch
			timeSeries[i] = models.TradeRecord{
				Symbol: trade.Symbol, Price: p,
				Time: t, Volume: v,
			}

			symbolTradeCounts[trade.Symbol]++
			if t.After(latestTimestamps[trade.Symbol]) {
				latestTimestamps[trade.Symbol] = t
			}
		}

		// Insert trade records in batch
		_, err = tradeCollection.InsertMany(context.Background(), timeSeries)
		if err != nil {
			log.Printf("Failed to batch insert trade records into MongoDB: %v", err)
		} else {
			log.Printf("Successfully inserted %d trade records.", len(timeSeries))
		}
	}
}
