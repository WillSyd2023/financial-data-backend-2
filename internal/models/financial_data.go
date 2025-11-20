package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SymbolDocument struct {
	Id          primitive.ObjectID `bson:"_id,omitempty"`
	Symbol      string             `bson:"symbol"`
	TradeCount  int64              `bson:"tradeCount"`
	LastTradeAt time.Time          `bson:"lastTradeAt"`
}

type TradeRecord struct {
	Id         primitive.ObjectID   `bson:"_id,omitempty"`
	MessageKey string               `bson:"message_key"`
	Symbol     string               `bson:"symbol"`
	Price      primitive.Decimal128 `bson:"price"`
	Time       time.Time            `bson:"time"`
	Volume     primitive.Decimal128 `bson:"volume"`
}
