package dto

import "time"

// GetSymbols

type GetSymbolsSingle struct {
	Symbol      string    `json:"symbol"`
	TradeCount  int64     `json:"trade_count"`
	LastTradeAt time.Time `json:"last_trade_at"`
}

type GetSymbolsRes struct {
	Available []GetSymbolsSingle `json:"available"`
}

// GetTradesPerSymbol

type PaginatedTradesResponseDTO struct {
	Data       []TradeResponseDTO `json:"data"`
	Pagination PaginationDTO      `json:"pagination"`
}

type TradeResponseDTO struct {
	Timestamp string `json:"timestamp"`
	Price     string `json:"price"`
	Volume    string `json:"volume"`
}

type PaginationDTO struct {
	// A Unix millisecond timestamp. It will be null if there are no more pages.
	NextCursor *int64 `json:"next_cursor"`
}
