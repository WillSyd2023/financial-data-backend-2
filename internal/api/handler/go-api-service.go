package handler

import (
	"financial-data-backend-2/internal/api/dto"
	"financial-data-backend-2/internal/api/usecase"
	"net/http"

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
	symbols, err := hd.uc.GetSymbols(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	// process response before returning
	var GetSymbolsRes dto.GetSymbolsRes
	GetSymbolsRes.Available = make([]dto.GetSymbolsSingle,
		len(symbols))
	for i, symbol := range symbols {
		GetSymbolsRes.Available[i] =
			dto.GetSymbolsSingle{
				Symbol:      symbol.Symbol,
				TradeCount:  symbol.TradeCount,
				LastTradeAt: symbol.LastTradeAt,
			}
	}

	// return response
	ctx.JSON(http.StatusOK,
		gin.H{
			"message": nil,
			"error":   nil,
			"data":    GetSymbolsRes,
		})
}

func (hd *Handler) GetTradesPerSymbol(ctx *gin.Context) {
	// usecase
}
