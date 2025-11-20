package processor

import (
	"financial-data-backend-2/internal/models"
	"testing"

	kafkaGo "github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
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
