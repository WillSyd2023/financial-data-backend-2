package main

import (
	"context"
	"financial-data-backend-2/internal/config"
	mongoGo "financial-data-backend-2/internal/mongo"
	"log"
	"time"
)

func main() {
	// - Load Configuration
	cfg, err := config.LoadConfig("config/config.yml")
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

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
}
