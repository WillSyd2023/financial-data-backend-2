package usecase

import (
	"financial-data-backend-2/internal/api/repo"
	"financial-data-backend-2/internal/models"

	"github.com/gin-gonic/gin"
)

type UsecaseItf interface {
	GetSymbols(*gin.Context) ([]models.SymbolDocument, error)
}

type Usecase struct {
	rp repo.RepoItf
}

func NewUsecase(rp repo.RepoItf) *Usecase {
	return &Usecase{rp: rp}
}

func (uc *Usecase) GetSymbols(ctx *gin.Context) ([]models.SymbolDocument, error) {
	// repo
	return uc.rp.GetSymbols(ctx)
}
