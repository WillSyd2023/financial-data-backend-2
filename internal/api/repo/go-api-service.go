package repo

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

type RepoItf interface {
	GetSymbols(*gin.Context) error
}

type Repo struct {
	sc *mongo.Collection
	tc *mongo.Collection
}

func NewRepo(symbolCollection, tradeCollection *mongo.Collection) *Repo {
	return &Repo{sc: symbolCollection, tc: tradeCollection}
}

func (rp *Repo) GetSymbols(ctx *gin.Context) error {
	return nil
}
