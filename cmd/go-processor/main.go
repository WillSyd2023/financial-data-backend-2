package main

import (
	"context"
	"encoding/json"
	"errors"
	"financial-data-backend-2/internal/config"
	"financial-data-backend-2/internal/kafka"
	"financial-data-backend-2/internal/models"
	mongoGo "financial-data-backend-2/internal/mongo"
	"fmt"
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

	// Also ensure a unique index exists on the trade collection for message.
	// This is a one-time setup.
	_, err = tradeCollection.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys:    bson.M{"message_key": 1},
			Options: options.Index().SetUnique(true),
		},
	)
	if err != nil {
		log.Printf("Could not create unique index on message key (may already exist): %v", err)
	}

	// - The Read Loop
	log.Println("Waiting for messages...")
	for {
		// Read a message from Kafka
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Error reading message: %v", err)
			continue
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
		log.Printf("Unmarshalled Data: %+v", finnMsg)

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
		timeSeries := make([]any, 0)
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
			timeSeries = append(timeSeries, models.TradeRecord{
				Id: primitive.NewObjectID(),
				// idempotency key to prevent redundant insertion of data from MQ
				MessageKey: fmt.Sprintf("%s-%d-%d-%s-%d-%d", m.Topic, m.Partition, m.Offset,
					trade.Symbol, trade.Timestamp, i),
				Symbol: trade.Symbol,
				Price:  p,
				Time:   t,
				Volume: v,
			})

			symbolTradeCounts[trade.Symbol]++
			if t.After(latestTimestamps[trade.Symbol]) {
				latestTimestamps[trade.Symbol] = t
			}
		}
		if len(timeSeries) == 0 {
			continue
		}

		// Insert trade records in batch
		insertCtx, insertCancel := context.WithTimeout(context.Background(), 10*time.Second)
		_, err = tradeCollection.InsertMany(insertCtx, timeSeries)
		insertCancel()
		if err != nil {
			// Check if the error is a duplicate key error. This requires checking the error code.
			isDuplicateKeyError := false
			var e mongo.BulkWriteException
			if errors.As(err, &e) {
				for _, we := range e.WriteErrors {
					if we.Code == 11000 {
						isDuplicateKeyError = true
						break
					}
				}
			}

			if isDuplicateKeyError {
				// This is a "good" error. It means we've successfully prevented duplicates.
				// We'll log it and proceed to the metadata update, just in case the
				// first attempt failed before it could update the metadata.
				log.Print("Info: Blocked duplicate trade insertion for message")
			} else {
				// This is a "bad" error (DB down, etc.). We should not proceed.
				log.Printf("CRITICAL: Failed to insert trade records: %v. Skipping metadata update.", err)
				continue
			}
		} else {
			log.Printf("Successfully inserted %d trade records.", len(timeSeries))
		}

		// Update symbol metadata
		updateCtx, updateCancel := context.WithTimeout(context.Background(), 10*time.Second)
		upsertOpts := options.Update().SetUpsert(true)
		for symbol, count := range symbolTradeCounts {
			filter := bson.M{"symbol": symbol}

			// $inc increments the tradeCount by the number of trades in this message
			// $set updates the last trade time (or sets it on insert)
			update := bson.M{
				"$inc":         bson.M{"tradeCount": count},
				"$max":         bson.M{"lastTradeAt": latestTimestamps[symbol]},
				"$setOnInsert": bson.M{"symbol": symbol},
			}

			// Perform the upsert operation for each symbol
			_, err := symbolCollection.UpdateOne(updateCtx, filter, update, upsertOpts)
			if err != nil {
				// This is a non-critical failure. We log it but don't stop the system.
				// This is a "eventual consistency" trade-off.
				log.Printf("Failed to upsert symbol metadata for '%s': %v", symbol, err)
			}
		}
		updateCancel()

		log.Printf("Updated metadata for %d unique symbol(s).", len(symbolTradeCounts))
	}
}
