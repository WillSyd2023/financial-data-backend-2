package main

import (
	"context"
	"financial-data-backend-2/internal/api/middleware"
	"financial-data-backend-2/internal/config"
	mongoGo "financial-data-backend-2/internal/mongo"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
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

	// Setup server and middlewares
	r := gin.New()
	r.Use(gin.Logger())
	// r.Use(middleware.Error())
	r.Use(middleware.Timeout(10 * time.Second))

	// Setup apps

	// Endpoints:

	// Run server
	srv := &http.Server{
		Addr:    os.Getenv("SERVER_PORT"),
		Handler: r.Handler(),
	}
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %s\n", err)
	}
}
