package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/tempizhere/gofemarket/internal/model"
	"github.com/tempizhere/gofemarket/internal/repository/mocks"
)

// TestBalanceRepository_GetBalance проверяет получение баланса пользователя через репозиторий.
func TestBalanceRepository_GetBalance(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		userID      int
		setupMock   func(*mocks.MockBalanceRepository)
		wantBalance model.Balance
		wantErr     error
	}{
		{
			name:   "success",
			userID: 1,
			setupMock: func(mock *mocks.MockBalanceRepository) {
				mock.EXPECT().
					GetBalance(gomock.Any(), 1).
					Return(model.Balance{Current: 100.50, Withdrawn: 50.25}, nil)
			},
			wantBalance: model.Balance{Current: 100.50, Withdrawn: 50.25},
			wantErr:     nil,
		},
		{
			name:   "no balance",
			userID: 2,
			setupMock: func(mock *mocks.MockBalanceRepository) {
				mock.EXPECT().
					GetBalance(gomock.Any(), 2).
					Return(model.Balance{}, nil)
			},
			wantBalance: model.Balance{},
			wantErr:     nil,
		},
		{
			name:   "db error",
			userID: 3,
			setupMock: func(mock *mocks.MockBalanceRepository) {
				mock.EXPECT().
					GetBalance(gomock.Any(), 3).
					Return(model.Balance{}, errors.New("db error"))
			},
			wantBalance: model.Balance{},
			wantErr:     errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockBalanceRepository(ctrl)
			tt.setupMock(mockRepo)

			balance, err := mockRepo.GetBalance(ctx, tt.userID)
			assert.Equal(t, tt.wantBalance, balance)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestBalanceRepository_WithdrawWithCheck проверяет корректность списания баллов с баланса пользователя с учетом различных сценариев.
func TestBalanceRepository_WithdrawWithCheck(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		userID      int
		orderNumber string
		sum         float64
		setupMock   func(*mocks.MockBalanceRepository)
		wantErr     error
	}{
		{
			name:        "success",
			userID:      1,
			orderNumber: "123",
			sum:         50.0,
			setupMock: func(mock *mocks.MockBalanceRepository) {
				mock.EXPECT().
					WithdrawWithCheck(gomock.Any(), 1, "123", 50.0).
					Return(nil)
			},
			wantErr: nil,
		},
		{
			name:        "insufficient funds",
			userID:      2,
			orderNumber: "123",
			sum:         150.0,
			setupMock: func(mock *mocks.MockBalanceRepository) {
				mock.EXPECT().
					WithdrawWithCheck(gomock.Any(), 2, "123", 150.0).
					Return(model.ErrInsufficientFunds)
			},
			wantErr: model.ErrInsufficientFunds,
		},
		{
			name:        "no balance",
			userID:      3,
			orderNumber: "123",
			sum:         50.0,
			setupMock: func(mock *mocks.MockBalanceRepository) {
				mock.EXPECT().
					WithdrawWithCheck(gomock.Any(), 3, "123", 50.0).
					Return(model.ErrInsufficientFunds)
			},
			wantErr: model.ErrInsufficientFunds,
		},
		{
			name:        "db error on update",
			userID:      4,
			orderNumber: "123",
			sum:         50.0,
			setupMock: func(mock *mocks.MockBalanceRepository) {
				mock.EXPECT().
					WithdrawWithCheck(gomock.Any(), 4, "123", 50.0).
					Return(errors.New("db error"))
			},
			wantErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockBalanceRepository(ctrl)
			tt.setupMock(mockRepo)

			err := mockRepo.WithdrawWithCheck(ctx, tt.userID, tt.orderNumber, tt.sum)
			if tt.wantErr != nil {
				if errors.Is(tt.wantErr, model.ErrInsufficientFunds) {
					assert.ErrorIs(t, err, tt.wantErr)
				} else {
					assert.EqualError(t, err, tt.wantErr.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestBalanceRepository_GetWithdrawals проверяет получение истории списаний пользователя.
func TestBalanceRepository_GetWithdrawals(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name            string
		userID          int
		setupMock       func(*mocks.MockBalanceRepository)
		wantWithdrawals []model.Withdrawal
		wantErr         error
	}{
		{
			name:   "success with data",
			userID: 1,
			setupMock: func(mock *mocks.MockBalanceRepository) {
				mock.EXPECT().
					GetWithdrawals(gomock.Any(), 1).
					Return([]model.Withdrawal{
						{Order: "123", Sum: 50.0, ProcessedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
						{Order: "456", Sum: 25.0, ProcessedAt: time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)},
					}, nil)
			},
			wantWithdrawals: []model.Withdrawal{
				{Order: "123", Sum: 50.0, ProcessedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
				{Order: "456", Sum: 25.0, ProcessedAt: time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)},
			},
			wantErr: nil,
		},
		{
			name:   "no withdrawals",
			userID: 2,
			setupMock: func(mock *mocks.MockBalanceRepository) {
				mock.EXPECT().
					GetWithdrawals(gomock.Any(), 2).
					Return([]model.Withdrawal{}, nil)
			},
			wantWithdrawals: []model.Withdrawal{},
			wantErr:         nil,
		},
		{
			name:   "db error",
			userID: 3,
			setupMock: func(mock *mocks.MockBalanceRepository) {
				mock.EXPECT().
					GetWithdrawals(gomock.Any(), 3).
					Return(nil, errors.New("db error"))
			},
			wantWithdrawals: nil,
			wantErr:         errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockBalanceRepository(ctrl)
			tt.setupMock(mockRepo)

			withdrawals, err := mockRepo.GetWithdrawals(ctx, tt.userID)
			assert.Equal(t, tt.wantWithdrawals, withdrawals)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
