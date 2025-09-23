package main

import (
	"encoding/json"
	"log"
	"net/url"

	"financial-data-backend-2/internal/config"

	"github.com/gorilla/websocket"
)

func main() {
	// Load Configuration
	cfg, err := config.LoadConfig("config/config.yml")
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// Establish WebSocket Connection
	u := url.URL{Scheme: "wss", Host: "ws.finnhub.io", RawQuery: "token=" + cfg.Finnhub.Token}
	// log.Printf("Connecting to %s", u.String())

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer conn.Close()
	log.Println("Successfully connected to Finnhub WebSocket")

	// FinnHub has given ping messages before
	conn.SetPingHandler(nil)

	// Subscribe to Symbols
	// Iterate over the symbols from your config file
	for _, symbol := range cfg.Symbols {
		subMsg, _ := json.Marshal(map[string]interface{}{"type": "subscribe", "symbol": symbol})
		log.Printf("Subscribing to %s", symbol)
		if err := conn.WriteMessage(websocket.TextMessage, subMsg); err != nil {
			log.Fatalf("Failed to subscribe to %s: %v", symbol, err)
		}
	}

	// The Read Loop (Run Forever)
	log.Println("Waiting for messages...")
	for {
		// Read a message from the connection
		var message interface{}
		err := conn.ReadJSON(&message)
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		msgMap, ok := message.(map[string]interface{})
		if ok && msgMap["type"] == "ping" {
			// Skip logging pings
			continue
		}

		// For now, just print the raw message to the console
		// This proves that the entire connection and subscription logic works.
		log.Printf("Received message: %+v", message)
	}
}
