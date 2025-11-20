package repo

import (
	"context"
	"financial-data-backend-2/internal/models"
	mongoGo "financial-data-backend-2/internal/mongo"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	databaseName          string = "financialDataDatabaseTest"
	symbolsCollectionName string = "symbols"
	tradesCollectionName  string = "finnhub_trades"
	testSymbol            string = "TEST"

	testRepo             *Repo
	testSymbolCollection *mongo.Collection
	testTradeCollection  *mongo.Collection

	// We'll create 20 trades, 1 second apart, with the most recent being 'now'.
	mockTradeData []any = make([]any, 20)
	now           time.Time
)

func TestMain(m *testing.M) {
	// 1. SETUP
	// Load URL
	err := godotenv.Load("../../../.env")
	if err != nil {
		log.Println("Warning: .env file not found, relying on environment variables for tests.")
		os.Exit(0)
	}
	mongoUrl := os.Getenv("MONGO_URL_TEST")
	if mongoUrl == "" {
		log.Println("Skipping repository tests: MONGO_URL_TEST environment variable not set.")
		os.Exit(0)
	}

	// Setup MongoDB collections and repo
	testDbClient, err := mongoGo.ConnectDB(mongoUrl)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	testSymbolCollection = testDbClient.Database(databaseName).Collection(symbolsCollectionName)
	testTradeCollection = testDbClient.Database(databaseName).Collection(tradesCollectionName)
	testRepo = NewRepo(testSymbolCollection, testTradeCollection)

	// Create our mock data
	now = time.Now().UTC().Truncate(time.Millisecond)
	for i := 0; i < 20; i++ {
		price, _ := primitive.ParseDecimal128("100.0")
		volume, _ := primitive.ParseDecimal128("10")
		mockTradeData[i] = models.TradeRecord{
			Id:         primitive.NewObjectID(),
			MessageKey: fmt.Sprintf("mock-key-%d", i),
			Symbol:     testSymbol,
			Time:       now.Add(time.Duration(-i) * time.Second),
			Price:      price,
			Volume:     volume,
		}
	}

	// 2. RUN THE TESTS
	exitCode := m.Run()

	// 3. TEARDOWN
	log.Println("Tearing down test database...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := testDbClient.Database(databaseName).Drop(ctx); err != nil {
		log.Fatalf("Error when tearing down test database: %v", err)
	}
	//
	log.Println("Test database torn.")
	if err := testDbClient.Disconnect(ctx); err != nil {
		log.Fatalf("Error during MongoDB disconnect: %v", err)
	}
	log.Println("MongoDB client disconnected.")

	os.Exit(exitCode)
}
func TestGetSymbols(t *testing.T) {
	testCases := []struct {
		name                string
		collectionInput     func() []any
		expectedNumSymbols  int
		expectedFirstSymbol string
	}{
		{
			name: "no symbol added",
			collectionInput: func() []any {
				return make([]any, 0)
			},
			expectedNumSymbols:  0,
			expectedFirstSymbol: "",
		},
		{
			name: "only one symbol added",
			collectionInput: func() []any {
				input := make([]any, 1)
				input[0] = models.SymbolDocument{
					Symbol: "A",
				}
				return input
			},
			expectedNumSymbols:  1,
			expectedFirstSymbol: "A",
		},
		{
			name: "several symbols added",
			collectionInput: func() []any {
				input := make([]any, 3)
				input[0] = models.SymbolDocument{
					Symbol: "Z",
				}
				input[1] = models.SymbolDocument{
					Symbol: "M",
				}
				input[2] = models.SymbolDocument{
					Symbol: "N",
				}
				return input
			},
			expectedNumSymbols:  3,
			expectedFirstSymbol: "M",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			// given
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_, err := testSymbolCollection.DeleteMany(ctx, bson.M{})
			assert.NoError(t, err)
			_, err = testSymbolCollection.InsertMany(ctx, tt.collectionInput())
			if tt.collectionInput() != nil && len(tt.collectionInput()) > 0 {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}

			// when
			symbols, err := testRepo.GetSymbols(context.Background())

			// then
			assert.NoError(t, err)
			if tt.collectionInput() != nil && len(tt.collectionInput()) > 0 {
				assert.NotNil(t, symbols)
			} else {
				assert.Nil(t, symbols)
			}
			assert.Len(t, symbols, tt.expectedNumSymbols)

			// If we expect results, check the first one to verify sorting
			if tt.expectedNumSymbols > 0 {
				assert.Equal(t, tt.expectedFirstSymbol, symbols[0].Symbol)
			}
		})
	}
}
func TestGetTradesPerSymbol(t *testing.T) {
	testCases := []struct {
		name                   string
		symbol                 string
		limit                  int
		before                 int64 // UnixMilli timestamp
		expectedNumTrades      int
		expectedFirstTradeTime time.Time
	}{
		{
			name:                   "Get first page (full page of 10)",
			symbol:                 testSymbol,
			limit:                  10,
			before:                 0, // No cursor, get the latest
			expectedNumTrades:      10,
			expectedFirstTradeTime: now, // The most recent trade
		},
		{
			name:                   "Get second page using cursor (full page of 10)",
			symbol:                 testSymbol,
			limit:                  10,
			before:                 now.Add(-9 * time.Second).UnixMilli(),
			expectedNumTrades:      10,
			expectedFirstTradeTime: now.Add(-10 * time.Second), // The 11th trade
		},
		{
			name:                   "Get partial last page",
			symbol:                 testSymbol,
			limit:                  10,
			before:                 now.Add(-14 * time.Second).UnixMilli(), // Cursor from 15th trade
			expectedNumTrades:      5,                                      // Should only get the remaining 5
			expectedFirstTradeTime: now.Add(-15 * time.Second),
		},
		{
			name:              "Non-existent symbol returns empty slice",
			symbol:            "NOSYMBOL",
			limit:             10,
			before:            0,
			expectedNumTrades: 0,
		},
		{
			name:              "Edge case: limit of 0 returns empty slice",
			symbol:            testSymbol,
			limit:             0,
			before:            0,
			expectedNumTrades: 0,
		},
		{
			name:              "No trades found after cursor",
			symbol:            testSymbol,
			limit:             10,
			before:            now.Add(-19 * time.Second).UnixMilli(), // Cursor from the very last trade
			expectedNumTrades: 0,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			// given
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_, err := testTradeCollection.DeleteMany(ctx, bson.M{})
			assert.NoError(t, err)
			_, err = testTradeCollection.InsertMany(ctx, mockTradeData)
			assert.NoError(t, err)

			// ACT
			trades, err := testRepo.GetTradesPerSymbol(context.Background(), tt.symbol, tt.limit, tt.before)

			// ASSERT
			assert.NoError(t, err)
			if tt.symbol == testSymbol && tt.expectedNumTrades > 0 {
				assert.NotNil(t, trades)
			} else {
				assert.Nil(t, trades)
			}
			assert.Len(t, trades, tt.expectedNumTrades)

			// If we expect results, check the first one to verify sorting
			if tt.expectedNumTrades > 0 {
				assert.Equal(t, tt.expectedFirstTradeTime, trades[0].Time)
			}
		})
	}
}
