package processor

import (
	"encoding/json"
	"errors"
	"financial-data-backend-2/internal/models"
	"fmt"
	"log"
	"strconv"
	"time"

	kafkaGo "github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ProcessedData struct {
	TradeRecords      []interface{}
	SymbolTradeCounts map[string]int64
	LatestTimestamps  map[string]time.Time
}

func IsDuplicateKeyError(err error) bool {
	var e mongo.BulkWriteException
	if errors.As(err, &e) {
		for _, we := range e.WriteErrors {
			if we.Code == 11000 {
				return true
			}
		}
	}
	return false
}

func TransformMessage(m kafkaGo.Message) (*ProcessedData, error) {
	// Unmarshal the raw JSON value from Kafka
	var finnMsg models.FinnhubTradeMessage
	if err := json.Unmarshal(m.Value, &finnMsg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	log.Printf("Unmarshalled Data: %+v", finnMsg)

	if finnMsg.Type != "trade" || len(finnMsg.Data) == 0 {
		return nil, nil // Not an error, just a message to skip (e.g., a ping)
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
		return nil, fmt.Errorf("message contained trade data, but all ticks were invalid")
	}

	return &ProcessedData{
		TradeRecords:      timeSeries,
		SymbolTradeCounts: symbolTradeCounts,
		LatestTimestamps:  latestTimestamps,
	}, nil
}
