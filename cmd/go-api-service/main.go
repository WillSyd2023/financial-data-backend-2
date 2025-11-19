package main

import (
	"context"
	"financial-data-backend-2/internal/api/handler"
	"financial-data-backend-2/internal/api/middleware"
	"financial-data-backend-2/internal/api/repo"
	"financial-data-backend-2/internal/api/usecase"
	"financial-data-backend-2/internal/config"
	mongoGo "financial-data-backend-2/internal/mongo"
	"log"
	"net/http"
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

	// - Setup collections
	sc := mongoGo.GetCollection(DB, cfg.MongoDB.DatabaseName,
		cfg.MongoDB.SymbolsCollectionName)
	tc := mongoGo.GetCollection(DB, cfg.MongoDB.DatabaseName,
		cfg.MongoDB.CollectionName)

	// Setup server and middlewares
	r := gin.New()
	r.Use(gin.Logger())
	// r.Use(middleware.Error())
	r.Use(middleware.Timeout(10 * time.Second))

	// Setup apps
	rp := repo.NewRepo(sc, tc)
	uc := usecase.NewUsecase(rp)
	hd := handler.NewHandler(uc)

	// Endpoints:
	v1 := r.Group("/api/v1")
	{
		// 1. Get metadata for all tracked symbols.
		v1.GET("/symbols", hd.GetSymbols)
		// 2. Get the 50 most recent trades for one symbol.
		v1.GET("/trades/:symbol", hd.GetTradesPerSymbol)
	}

	// Run server
	srv := &http.Server{
		Addr:    ":" + cfg.APIPort,
		Handler: r.Handler(),
	}
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %s\n", err)
	}
}
