package usecase

import (
	"context"
	"errors"
	"financial-data-backend-2/internal/api/repo"
	"financial-data-backend-2/internal/api/repo/mocks"
	"financial-data-backend-2/internal/models"
	"testing"
	"time"

	"github.com/go-playground/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestGetSymbols(t *testing.T) {
	testCases := []struct {
		name           string
		repoSetup      func(context.Context) repo.RepoItf
		expectedOutput func() []models.SymbolDocument
		expectedErr    error
	}{
		{
			name: "return empty array without error",
			repoSetup: func(ctx context.Context) repo.RepoItf {
				mock := new(mocks.RepoItf)
				var empty []models.SymbolDocument
				mock.On("GetSymbols", ctx).
					Return(empty, nil)
				return mock
			},
			expectedOutput: func() []models.SymbolDocument {
				var empty []models.SymbolDocument
				return empty
			},
			expectedErr: nil,
		}, {
			name: "return non-empty array without error",
			repoSetup: func(ctx context.Context) repo.RepoItf {
				mock := new(mocks.RepoItf)
				var nonempty []models.SymbolDocument
				nonempty = append(nonempty, models.SymbolDocument{
					Symbol:      "A",
					TradeCount:  14,
					LastTradeAt: time.UnixMilli(200),
				})
				mock.On("GetSymbols", ctx).
					Return(nonempty, nil)
				return mock
			},
			expectedOutput: func() []models.SymbolDocument {
				var nonempty []models.SymbolDocument
				nonempty = append(nonempty, models.SymbolDocument{
					Symbol:      "A",
					TradeCount:  14,
					LastTradeAt: time.UnixMilli(200),
				})
				return nonempty
			},
			expectedErr: nil,
		},
		{
			name: "return error",
			repoSetup: func(ctx context.Context) repo.RepoItf {
				mock := new(mocks.RepoItf)
				mock.On("GetSymbols", ctx).
					Return(nil, errors.New("api usecase error"))
				return mock
			},
			expectedOutput: func() []models.SymbolDocument {
				return nil
			},
			expectedErr: errors.New("api usecase error"),
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			//given
			uc := NewUsecase(tt.repoSetup(context.Background()))

			//when
			output, err := uc.GetSymbols(context.Background())

			//then
			assert.Equal(t, tt.expectedOutput(), output)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}
func TestGetTradesPerSymbol(t *testing.T) {
	price, _ := primitive.ParseDecimal128("123.50")
	volume, _ := primitive.ParseDecimal128("50")

	testCases := []struct {
		name           string
		inputSymbol    string
		inputLimit     int
		inputBefore    int64
		repoSetup      func(context.Context) repo.RepoItf
		expectedOutput func() []models.TradeRecord
		expectedErr    error
	}{
		{
			name:        "return empty array without error",
			inputSymbol: "A",
			inputLimit:  14,
			inputBefore: 256,
			repoSetup: func(ctx context.Context) repo.RepoItf {
				mock := new(mocks.RepoItf)
				var empty []models.TradeRecord
				mock.On("GetTradesPerSymbol", ctx, "A", 14, int64(256)).
					Return(empty, nil)
				return mock
			},
			expectedOutput: func() []models.TradeRecord {
				var empty []models.TradeRecord
				return empty
			},
			expectedErr: nil,
		},
		{
			name:        "return non-empty array without error",
			inputSymbol: "A",
			inputLimit:  14,
			inputBefore: 256,
			repoSetup: func(ctx context.Context) repo.RepoItf {
				var nonempty []models.TradeRecord
				nonempty = append(nonempty,
					models.TradeRecord{
						Symbol: "A",
						Price:  price,
						Time:   time.UnixMilli(200),
						Volume: volume,
					})
				mock := new(mocks.RepoItf)
				mock.On("GetTradesPerSymbol", ctx, "A", 14, int64(256)).
					Return(nonempty, nil)
				return mock
			},
			expectedOutput: func() []models.TradeRecord {
				var nonempty []models.TradeRecord
				nonempty = append(nonempty,
					models.TradeRecord{
						Symbol: "A",
						Price:  price,
						Time:   time.UnixMilli(200),
						Volume: volume,
					})
				return nonempty
			},
			expectedErr: nil,
		},
		{
			name:        "return error",
			inputSymbol: "A",
			inputLimit:  14,
			inputBefore: 256,
			repoSetup: func(ctx context.Context) repo.RepoItf {
				mock := new(mocks.RepoItf)
				mock.On("GetTradesPerSymbol", ctx, "A", 14, int64(256)).
					Return(nil, errors.New("api usecase error"))
				return mock
			},
			expectedOutput: func() []models.TradeRecord {
				return nil
			},
			expectedErr: errors.New("api usecase error"),
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			//given
			uc := NewUsecase(tt.repoSetup(context.Background()))

			//when
			output, err := uc.GetTradesPerSymbol(context.Background(),
				tt.inputSymbol, tt.inputLimit, tt.inputBefore)

			//then
			assert.Equal(t, tt.expectedOutput(), output)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}
