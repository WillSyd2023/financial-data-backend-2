package main

import (
	"encoding/json"
	"log"
	"net/url"

	"financial-data-backend-2/internal/config" // Your internal config package

	"github.com/gorilla/websocket" // The WebSocket library
)

func main() {
	// 1. Load Configuration
	cfg, err := config.LoadConfig("config/config.yml")
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// 2. Establish WebSocket Connection
	// Construct the URL with your API token
	u := url.URL{Scheme: "wss", Host: "ws.finnhub.io", RawQuery: "token=" + cfg.Finnhub.Token}
	// log.Printf("Connecting to %s", u.String())

	// Dial the connection
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	// Ensure the connection is closed when the function exits
	defer conn.Close()
	log.Println("Successfully connected to Finnhub WebSocket")

	conn.SetPingHandler(nil)

	// 3. Subscribe to Symbols
	// Iterate over the symbols from your config file
	for _, symbol := range cfg.Symbols {
		subMsg, _ := json.Marshal(map[string]interface{}{"type": "subscribe", "symbol": symbol})
		log.Printf("Subscribing to %s", symbol)
		// Write the subscription message as JSON to the WebSocket
		if err := conn.WriteMessage(websocket.TextMessage, subMsg); err != nil {
			log.Fatalf("Failed to subscribe to %s: %v", symbol, err)
		}
	}

	// 4. The Read Loop (Run Forever)
	// This is the core of the ingestor. It will block here and continuously
	// listen for incoming messages.
	log.Println("Waiting for messages...")
	for {
		// Read a message from the connection
		// We can read it as a generic map[string]interface{} for now
		var message interface{}
		err := conn.ReadJSON(&message)
		if err != nil {
			log.Printf("Error reading message: %v", err)
			// If there's an error, it's often best to break the loop and let the program exit/restart
			break
		}

		// The ping handler will catch ping messages, so we can add a check here
		// to be sure we don't log them if they slip through.
		msgMap, ok := message.(map[string]interface{})
		if ok && msgMap["type"] == "ping" {
			continue // Skip logging pings
		}

		// For now, just print the raw message to the console!
		// This proves that your entire connection and subscription logic works.
		log.Printf("Received message: %+v", message)
	}
}
