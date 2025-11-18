package usecase

import (
	"context"
	"financial-data-backend-2/internal/api/repo"
	"financial-data-backend-2/internal/models"
)

type UsecaseItf interface {
	GetSymbols(context.Context) ([]models.SymbolDocument, error)
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
