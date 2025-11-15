package models

type FinnhubTradeData struct {
	Price     float64 `json:"p"` // Last price
	Symbol    string  `json:"s"` // Symbol (e.g., "AAPL")
	Timestamp int64   `json:"t"` // Unix timestamp in milliseconds
	Volume    float64 `json:"v"` // Volume
}

type FinnhubTradeMessage struct {
	Type string             `json:"type"`
	Data []FinnhubTradeData `json:"data"` // List of trades or price updates
}
