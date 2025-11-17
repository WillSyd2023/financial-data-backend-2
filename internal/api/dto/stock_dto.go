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
