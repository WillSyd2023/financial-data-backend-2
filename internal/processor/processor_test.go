package processor

import (
	"context"
	"financial-data-backend-2/internal/models"
	mongoGo "financial-data-backend-2/internal/mongo"
	"log"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	kafkaGo "github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestTransformMessage(t *testing.T) {
	testCases := []struct {
		name          string
		inputMessage  kafkaGo.Message
		expectError   bool
		expectNilData bool                                    // For cases that should be skipped (nil, nil)
		assertions    func(t *testing.T, data *ProcessedData) // Custom checks for success cases
	}{
		{
			name:          "JSON unmarshal failure",
			inputMessage:  kafkaGo.Message{Value: []byte(`{"type":"trade","data":[{"s":"AAPL","p":"not_a_number","v":100,"t":1678886400123}]}`)},
			expectError:   true,
			expectNilData: true,
		},
		{
			name:          "should skip non-trade messages (e.g., pings)",
			inputMessage:  kafkaGo.Message{Value: []byte(`{"type":"ping"}`)},
			expectError:   false,
			expectNilData: true,
		},
		{
			name:          "should skip messages with an empty data array",
			inputMessage:  kafkaGo.Message{Value: []byte(`{"type":"trade", "data":[]}`)},
			expectError:   false,
			expectNilData: true,
		},
		{
			name: "should successfully transform a valid trade message",
			inputMessage: kafkaGo.Message{
				Topic:     "test-topic",
				Partition: 1,
				Offset:    42,
				Value:     []byte(`{"type":"trade","data":[{"s":"AAPL","p":150.75,"v":100.5,"t":1678886400123}]}`),
			},
			expectError:   false,
			expectNilData: false,
			assertions: func(t *testing.T, data *ProcessedData) {
				assert.Len(t, data.TradeRecords, 1)
				record := data.TradeRecords[0].(models.TradeRecord)

				// Check data transformation
				assert.False(t, record.Id.IsZero(), "A new ObjectID should have been generated")
				assert.Equal(t, "test-topic-1-42-AAPL-1678886400123-0", record.MessageKey)
				assert.Equal(t, "AAPL", record.Symbol)
				assert.Equal(t, "150.75", record.Price.String())
				assert.Equal(t, int64(1678886400123), record.Time.UnixMilli())
				assert.Equal(t, "100.5", record.Volume.String())

				// Check metadata maps
				assert.Len(t, data.SymbolTradeCounts, 1)
				assert.Equal(t, int64(1), data.SymbolTradeCounts["AAPL"])
				assert.Len(t, data.LatestTimestamps, 1)
				assert.Equal(t, int64(1678886400123), data.LatestTimestamps["AAPL"].UnixMilli())
			},
		},
		{
			name: "message keys should be unique and deterministic",
			inputMessage: kafkaGo.Message{
				Topic:     "key-topic",
				Partition: 2,
				Offset:    101,
				Value: []byte(`{"type":"trade","data":[
					{"s":"MSFT","p":300,"v":10,"t":1700000000000},
					{"s":"MSFT","p":301,"v":20,"t":1700000000000}
				]}`),
			},
			expectError:   false,
			expectNilData: false,
			assertions: func(t *testing.T, data *ProcessedData) {
				assert.Len(t, data.TradeRecords, 2)
				record1 := data.TradeRecords[0].(models.TradeRecord)
				record2 := data.TradeRecords[1].(models.TradeRecord)

				// Test for UNIQUENESS
				assert.NotEqual(t, record1.MessageKey, record2.MessageKey)

				// Test for DETERMINISM (by checking the expected format)
				assert.Equal(t, "key-topic-2-101-MSFT-1700000000000-0", record1.MessageKey)
				assert.Equal(t, "key-topic-2-101-MSFT-1700000000000-1", record2.MessageKey)
			},
		},
	}

	// --- RUNNER: Loop through all test cases ---
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			processedData, err := TransformMessage(tt.inputMessage)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.expectNilData {
				assert.Nil(t, processedData)
			} else {
				assert.NotNil(t, processedData)
			}

			// If we have custom assertions for a success case, run them.
			if tt.assertions != nil {
				tt.assertions(t, processedData)
			}
		})
	}
}

var (
	databaseName         string = "financialDataDatabaseTest"
	tradesCollectionName string = "finnhub_trades"
)

func TestInsertMany_Idempotency(t *testing.T) {
	/**
	Tests that idempotency-key approach on go-processor main
	via unique message key works as intended via demoing,
	but not testing actual go-processor code.
	**/
	// --- ARRANGE ---
	// 1. Setup MongoDB for use
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Println("Warning: .env file not found, relying on environment variables for tests.")
		os.Exit(0)
	}
	mongoUrl := os.Getenv("MONGO_URL_TEST")
	if mongoUrl == "" {
		log.Println("Skipping repository tests: MONGO_URL_TEST environment variable not set.")
		os.Exit(0)
	}
	testDbClient, err := mongoGo.ConnectDB(mongoUrl, 15*time.Second)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	testTradeCollection := testDbClient.Database(databaseName).Collection(tradesCollectionName)

	// 2. Define a predictable batch of data with deterministic keys
	testTimestamp := time.Now().UTC()
	price, _ := primitive.ParseDecimal128("100")
	volume, _ := primitive.ParseDecimal128("10")

	idempotentBatch := []interface{}{
		models.TradeRecord{
			Id:         primitive.NewObjectID(),
			MessageKey: "test-topic-0-1-TEST-12345-0", // Manually created key
			Symbol:     "TEST",
			Price:      price,
			Time:       testTimestamp,
			Volume:     volume,
		},
		models.TradeRecord{
			Id:         primitive.NewObjectID(),
			MessageKey: "test-topic-0-1-TEST-12345-1", // Manually created key
			Symbol:     "TEST",
			Price:      price,
			Time:       testTimestamp.Add(1 * time.Second),
			Volume:     volume,
		},
	}

	// 3. Ensure the test collection is in a clean state before we start.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = testTradeCollection.Drop(ctx)
	assert.NoError(t, err)

	// Also ensure a unique index exists on the trade collection for message.
	// This is a one-time setup.
	_, err = testTradeCollection.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys:    bson.M{"message_key": 1},
			Options: options.Index().SetUnique(true),
		},
	)
	assert.NoError(t, err)

	// --- ACT 1: The first insert ---
	_, err = testTradeCollection.InsertMany(ctx, idempotentBatch)

	// --- ASSERT 1: The first insert must succeed ---
	assert.NoError(t, err, "The first insert should succeed without any errors")

	// Verify that the data is actually there
	count, err := testTradeCollection.CountDocuments(ctx, bson.M{"symbol": "TEST"})
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count, "There should be 2 documents in the collection after the first insert")

	// --- ACT 2: The second insert (the idempotency test) ---
	_, err = testTradeCollection.InsertMany(ctx, idempotentBatch)

	// --- ASSERT 2: The second insert MUST fail with the correct error ---
	assert.Error(t, err, "The second insert should fail")

	// Check if the error is the specific DuplicateKeyError we expect.
	assert.True(t, IsDuplicateKeyError(err), "The error should be a duplicate key error")

	// The ultimate proof: Verify that no new data was written.
	finalCount, err := testTradeCollection.CountDocuments(ctx, bson.M{"symbol": "TEST"})
	assert.NoError(t, err)
	assert.Equal(t, int64(2), finalCount, "The document count should still be 2 after the failed second insert")
}
