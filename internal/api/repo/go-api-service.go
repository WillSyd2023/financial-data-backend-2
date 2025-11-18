package repo

import (
	"context"
	"financial-data-backend-2/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

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

func (rp *Repo) GetTradesPerSymbol(ctx context.Context, symbol string, limit int, before int64) ([]models.TradeRecord, error) {
	return nil, nil
}
