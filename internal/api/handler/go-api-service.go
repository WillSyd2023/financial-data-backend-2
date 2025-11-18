package handler

import (
	"financial-data-backend-2/internal/api/constant"
	"financial-data-backend-2/internal/api/dto"
	"financial-data-backend-2/internal/api/usecase"
	"financial-data-backend-2/internal/models"
	"net/http"
	"strconv"

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
	// request validation
	symbol := ctx.Param("symbol")
	if symbol == "" {
		ctx.Error(constant.ErrNoSymbol)
		return
	}

	// Get limit and offset
	var limit int
	limitStr := ctx.Query("limit")
	if limitStr == "" {
		limit = constant.DefaultLimit
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = constant.DefaultLimit
	}
	var offset int
	offsetStr := ctx.Query("offset")
	if offsetStr == "" {
		offset = constant.DefaultOffset
	}
	offset, err = strconv.Atoi(offsetStr)
	if err != nil {
		offset = constant.DefaultOffset
	}

	params := models.GetTradesPerSymbolParams{
		Symbol: symbol,
		PaginationQuery: models.PaginationQuery{
			Limit: limit, Offset: offset,
		},
	}
}
