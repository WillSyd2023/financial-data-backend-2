package usecase

import (
	"financial-data-backend-2/internal/api/repo"

	"github.com/gin-gonic/gin"
)

type UsecaseItf interface {
	GetSymbols(*gin.Context) error
}

type Usecase struct {
	rp repo.RepoItf
}

func NewUsecase(rp repo.RepoItf) *Usecase {
	return &Usecase{rp: rp}
}

func (uc *Usecase) GetSymbols(ctx *gin.Context) error {
	// repo
	return nil
}
