package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SymbolDocument struct {
	Id          primitive.ObjectID `bson:"_id,omitempty"`
	Symbol      string             `bson:"symbol"`
	TradeCount  int64              `bson:"trade_count"`
	LastTradeAt time.Time          `bson:"last_trade_at"`
}

type FinnHubTradeRecord struct {
	Id     primitive.ObjectID   `bson:"_id,omitempty"`
	Symbol string               `bson:"symbol"`
	Price  primitive.Decimal128 `bson:"price"`
	Time   time.Time            `bson:"time"`
	Volume int64                `bson:"volume"`
}
