package service

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

// TestBalanceService_GetBalance проверяет получение баланса пользователя через сервис.
func TestBalanceService_GetBalance(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	repo := mocks.NewMockBalanceRepository(ctrl)
	s := NewBalanceService(repo, nil).(*balanceServiceImpl)

	expectedBalance := model.Balance{Current: 100.0, Withdrawn: 50.0}
	repo.EXPECT().GetBalance(ctx, 1).Return(expectedBalance, nil).Times(1)

	balance, err := s.GetBalance(ctx, 1)
	assert.NoError(t, err)
	assert.Equal(t, expectedBalance, balance)
}

// TestBalanceService_Withdraw проверяет вывод средств с учетом формата заказа и баланса.
func TestBalanceService_Withdraw(t *testing.T) {
	tests := []struct {
		name        string
		orderNumber string
		userID      int
		sum         float64
		repoErr     error
		wantErr     error
	}{
		{
			name:        "valid withdrawal",
			orderNumber: "12345678903",
			userID:      1,
			sum:         50.0,
			repoErr:     nil,
			wantErr:     nil,
		},
		{
			name:        "invalid order format",
			orderNumber: "123",
			userID:      1,
			sum:         50.0,
			repoErr:     nil,
			wantErr:     model.ErrInvalidOrderFormat,
		},
		{
			name:        "repo error",
			orderNumber: "12345678903",
			userID:      1,
			sum:         50.0,
			repoErr:     model.ErrInsufficientFunds,
			wantErr:     model.ErrInsufficientFunds,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := context.Background()
			repo := mocks.NewMockBalanceRepository(ctrl)
			s := NewBalanceService(repo, nil).(*balanceServiceImpl)

			if tt.wantErr != model.ErrInvalidOrderFormat {
				repo.EXPECT().WithdrawWithCheck(ctx, tt.userID, tt.orderNumber, tt.sum).Return(tt.repoErr).Times(1)
			}
			err := s.Withdraw(ctx, tt.userID, tt.orderNumber, tt.sum)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestBalanceService_GetWithdrawals проверяет получение истории выводов пользователя через сервис.
func TestBalanceService_GetWithdrawals(t *testing.T) {
	t.Parallel()

	// success case
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		repo := mocks.NewMockBalanceRepository(ctrl)
		s := NewBalanceService(repo, nil).(*balanceServiceImpl)

		fixedTime := time.Date(2025, time.July, 2, 10, 0, 0, 0, time.UTC)
		expectedWithdrawals := []model.Withdrawal{{Order: "12345678903", Sum: 50.0, ProcessedAt: fixedTime}}
		repo.EXPECT().GetWithdrawals(ctx, 1).Return(expectedWithdrawals, nil).Times(1)

		withdrawals, err := s.GetWithdrawals(ctx, 1)
		assert.NoError(t, err)
		assert.Equal(t, expectedWithdrawals, withdrawals)
	})

	// error case
	t.Run("error case", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		repo := mocks.NewMockBalanceRepository(ctrl)
		s := NewBalanceService(repo, nil).(*balanceServiceImpl)

		repo.EXPECT().GetWithdrawals(ctx, 1).Return(nil, errors.New("db error")).Times(1)
		_, err := s.GetWithdrawals(ctx, 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
	})
}
