package repo

import (
	"context"
	"financial-data-backend-2/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//go:generate mockery --name RepoItf --case underscore --keeptree
type RepoItf interface {
	GetSymbols(context.Context) ([]models.SymbolDocument, error)
	GetTradesPerSymbol(context.Context, string, int, int64) ([]models.TradeRecord, error)
}

type Repo struct {
	sc *mongo.Collection
	tc *mongo.Collection
}

func NewRepo(symbolCollection, tradeCollection *mongo.Collection) *Repo {
	return &Repo{sc: symbolCollection, tc: tradeCollection}
}

func (rp *Repo) GetSymbols(c context.Context) ([]models.SymbolDocument, error) {
	results, err := rp.sc.Find(c, bson.D{}, options.Find().SetSort(
		bson.D{{Key: "symbol", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer results.Close(c)

	var symbols []models.SymbolDocument
	if err = results.All(c, &symbols); err != nil {
		return nil, err
	}
	return symbols, nil
}

func (r *Repo) GetTradesPerSymbol(ctx context.Context, symbol string, limit int, before int64) ([]models.TradeRecord, error) {
	var trades []models.TradeRecord
	if limit <= 0 {
		return nil, nil
	}

	filter := bson.M{"symbol": symbol}

	// If a 'before' cursor is provided, add it to the filter.
	// This finds all trades OLDER than the cursor.
	if before > 0 {
		filter["time"] = bson.M{"$lt": time.UnixMilli(before)}
	}

	findOptions := options.Find().
		SetSort(bson.D{{Key: "time", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := r.tc.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &trades); err != nil {
		return nil, err
	}

	return trades, nil
}
