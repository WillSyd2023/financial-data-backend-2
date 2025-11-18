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

	symbols := make([]models.SymbolDocument, 0)
	defer results.Close(c)
	for results.Next(c) {
		var symbol models.SymbolDocument
		if err = results.Decode(&symbol); err != nil {
			return nil, err
		}
		symbols = append(symbols, symbol)
	}

	return symbols, nil
}
