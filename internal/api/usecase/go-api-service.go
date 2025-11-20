package usecase

import (
	"context"
	"financial-data-backend-2/internal/api/repo"
	"financial-data-backend-2/internal/models"
)

//go:generate mockery --name UsecaseItf --case underscore --keeptree
type UsecaseItf interface {
	GetSymbols(context.Context) ([]models.SymbolDocument, error)
	GetTradesPerSymbol(context.Context, string, int, int64) ([]models.TradeRecord, error)
}

type Usecase struct {
	rp repo.RepoItf
}

func NewUsecase(rp repo.RepoItf) *Usecase {
	return &Usecase{rp: rp}
}

func (uc *Usecase) GetSymbols(ctx context.Context) ([]models.SymbolDocument, error) {
	// repo
	return uc.rp.GetSymbols(ctx)
}

func (uc *Usecase) GetTradesPerSymbol(ctx context.Context, symbol string, limit int, before int64) ([]models.TradeRecord, error) {
	// repo
	return uc.rp.GetTradesPerSymbol(ctx, symbol, limit, before)
}
