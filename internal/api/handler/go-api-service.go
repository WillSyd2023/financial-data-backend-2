package handler

import (
	"financial-data-backend-2/internal/api/usecase"

	"github.com/gin-gonic/gin"
)

type HandlerItf interface {
	GetSymbols(*gin.Context)
	GetTradesPerSymbol(*gin.Context)
}

type Handler struct {
	uc usecase.UsecaseItf
}

func NewHandler(uc usecase.UsecaseItf) *Handler {
	return &Handler{uc: uc}
}

func (hd *Handler) GetSymbols(ctx *gin.Context) {
	// usecase
}

func (hd *Handler) GetTradesPerSymbol(ctx *gin.Context) {
	// usecase
}
