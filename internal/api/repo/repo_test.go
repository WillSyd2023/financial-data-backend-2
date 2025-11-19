package repo

import (
	"context"
	"financial-data-backend-2/internal/models"
	mongoGo "financial-data-backend-2/internal/mongo"
	"log"
	"os"
	"testing"
	"time"

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
			Id:     primitive.NewObjectID(),
			Symbol: testSymbol,
			Time:   now.Add(time.Duration(-i) * time.Second),
			Price:  price,
			Volume: volume,
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
	log.Println("Test database torn.")
	if err := testDbClient.Disconnect(ctx); err != nil {
		log.Fatalf("Error during MongoDB disconnect: %v", err)
	}
	log.Println("MongoDB client disconnected.")

	os.Exit(exitCode)
}
