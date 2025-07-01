package service

import (
	"context"

	"github.com/tempizhere/gofemarket/internal/api"
	"github.com/tempizhere/gofemarket/internal/model"
	"github.com/tempizhere/gofemarket/internal/repository"
)

// balanceServiceImpl реализует логику баланса.
type balanceServiceImpl struct {
	balanceRepo *repository.BalanceRepository
}

// NewBalanceService создает новый BalanceService.
func NewBalanceService(balanceRepo *repository.BalanceRepository, _ *repository.OrderRepository) api.BalanceService {
	return &balanceServiceImpl{
		balanceRepo: balanceRepo,
	}
}

// GetBalance возвращает текущий баланс.
func (s *balanceServiceImpl) GetBalance(ctx context.Context, userID int) (model.Balance, error) {
	return s.balanceRepo.GetBalance(ctx, userID)
}

// Withdraw выполняет списание баллов.
func (s *balanceServiceImpl) Withdraw(ctx context.Context, userID int, orderNumber string, sum float64) error {
	if !isValidLuhn(orderNumber) {
		return model.ErrInvalidOrderFormat
	}

	return s.balanceRepo.WithdrawWithCheck(ctx, userID, orderNumber, sum)
}

// GetWithdrawals возвращает историю списаний.
func (s *balanceServiceImpl) GetWithdrawals(ctx context.Context, userID int) ([]model.Withdrawal, error) {
	return s.balanceRepo.GetWithdrawals(ctx, userID)
}
