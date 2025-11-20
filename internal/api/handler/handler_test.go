package handler

import (
	"errors"
	"financial-data-backend-2/internal/api/constant"
	"financial-data-backend-2/internal/api/middleware"
	"financial-data-backend-2/internal/api/usecase"
	"financial-data-backend-2/internal/api/usecase/mocks"
	"financial-data-backend-2/internal/models"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupRouter(uc usecase.UsecaseItf) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Apply the REAL middlewares we want to test
	r.Use(middleware.Error())
	r.Use(middleware.Timeout(100 * time.Millisecond)) // A short timeout for testing

	handler := NewHandler(uc)

	// Register the real routes
	v1 := r.Group("/api/v1")
	{
		v1.GET("/symbols", handler.GetSymbols)
		v1.GET("/trades/:symbol", handler.GetTradesPerSymbol)
	}
	return r
}
func TestIntegratedGetTradesPerSymbolHandler(t *testing.T) {
	/**
	Instead of testing the handler logic thoroughly, these are
	mostly testing integration of handler with error and timeout middleware
	and (mock) usecase, though the following do test logic a bit:
	- "Success - should return trades with correct DTO format"
	- "Failure - invalid limit parameter (returns custom error)"
	**/
	mockTradeTime := time.Now()
	mockTrades := []models.TradeRecord{
		{Time: mockTradeTime},
	}
	usecaseError := errors.New("a simulated usecase error")

	testCases := []struct {
		name                 string
		url                  string
		setupMock            func(mockUC *mocks.UsecaseItf)
		expectedStatusCode   int
		expectedBodyContains string
	}{
		{
			name: "Success - should return trades with correct DTO format",
			url:  "/api/v1/trades/AAPL?limit=1",
			setupMock: func(mockUC *mocks.UsecaseItf) {
				// We expect the handler to parse "AAPL", 1, and 0 and pass them here.
				mockUC.On("GetTradesPerSymbol", mock.Anything, "AAPL", 1, int64(0)).Return(mockTrades, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedBodyContains: `"next_cursor":` + fmt.Sprintf("%d", mockTradeTime.UnixMilli()),
		},
		{
			name: "Failure - invalid limit parameter (returns custom error)",
			url:  "/api/v1/trades/AAPL?limit=abc",
			setupMock: func(mockUC *mocks.UsecaseItf) {
				// The usecase should NOT be called if parameter parsing fails.
			},
			expectedStatusCode:   http.StatusBadRequest,
			expectedBodyContains: "invalid 'limit' query parameter: must be a positive integer",
		},
		{
			name: "Failure - usecase returns a custom error",
			url:  "/api/v1/trades/TSLA",
			setupMock: func(mockUC *mocks.UsecaseItf) {
				mockUC.On("GetTradesPerSymbol", mock.Anything, "TSLA", constant.DefaultLimit, int64(0)).Return(nil, constant.ErrNoSymbol)
			},
			expectedStatusCode:   http.StatusBadRequest,
			expectedBodyContains: constant.ErrNoSymbol.Error(),
		},
		{
			name: "Failure - usecase returns a generic error",
			url:  "/api/v1/trades/NVDA",
			setupMock: func(mockUC *mocks.UsecaseItf) {
				mockUC.On("GetTradesPerSymbol", mock.Anything, "NVDA", constant.DefaultLimit, int64(0)).Return(nil, usecaseError)
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedBodyContains: usecaseError.Error(),
		},
		{
			name: "Failure - usecase is too slow and times out",
			url:  "/api/v1/trades/GOOGL",
			setupMock: func(mockUC *mocks.UsecaseItf) {
				mockUC.On("GetTradesPerSymbol", mock.Anything, "GOOGL", constant.DefaultLimit, int64(0)).
					// This mock will sleep for longer than the middleware timeout.
					After(200*time.Millisecond).
					Return(nil, nil)
			},
			expectedStatusCode:   http.StatusGatewayTimeout,
			expectedBodyContains: "request timed out",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			// ARRANGE
			mockUC := new(mocks.UsecaseItf)
			tt.setupMock(mockUC)
			router := setupRouter(mockUC)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tt.url, nil)

			// ACT
			router.ServeHTTP(w, req)

			// ASSERT
			assert.Equal(t, tt.expectedStatusCode, w.Code, "status code should match")
			assert.Contains(t, w.Body.String(), tt.expectedBodyContains, "response body should contain expected text")

			// Verify that the mock was called as expected.
			mockUC.AssertExpectations(t)
		})
	}
}
