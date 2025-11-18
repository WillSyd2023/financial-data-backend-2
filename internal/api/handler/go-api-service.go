package handler

import (
	"financial-data-backend-2/internal/api/constant"
	"financial-data-backend-2/internal/api/dto"
	"financial-data-backend-2/internal/api/usecase"
	"net/http"
	"strconv"
	"time"

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
	symbols, err := hd.uc.GetSymbols(ctx.Request.Context())
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
	if err != nil || limit <= 0 {
		limit = constant.DefaultLimit
	}

	// Parse the 'before' cursor. It's a Unix millisecond timestamp.
	// If it's not provided, it defaults to 0, which our service
	// will treat as "get the latest".
	var before int64
	beforeStr := ctx.DefaultQuery("before", constant.DefaultCursorStr)
	if beforeStr == "" {
		before = constant.DefaultCursor
	}
	before, err = strconv.ParseInt(beforeStr, 10, 64)
	if err != nil || before < 0 {
		before = constant.DefaultCursor
	}

	// usecase
	trades, err := hd.uc.GetTradesPerSymbol(ctx.Request.Context(),
		symbol, limit, before)
	if err != nil {
		ctx.Error(err)
		return
	}

	// Figure out response DTO
	// - trades
	var res dto.PaginatedTradesResponseDTO
	for _, trade := range trades {
		res.Data = append(res.Data,
			dto.TradeResponseDTO{
				Timestamp: trade.Time.Format(time.RFC3339Nano),
				Price:     trade.Price.String(),
				Volume:    trade.Volume.String(),
			})
	}

	// - next cursor
	res.Pagination = dto.PaginationDTO{}
	if len(trades) == limit {
		next := trades[len(trades)-1].Time.UnixMilli()
		res.Pagination.NextCursor = &next
	}

	// return response
	ctx.JSON(http.StatusOK,
		gin.H{
			"message": nil,
			"error":   nil,
			"data":    res,
		})
}
