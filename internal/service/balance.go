package service

import (
	"context"

	"github.com/tempizhere/gofemarket/internal/model"
	"github.com/tempizhere/gofemarket/internal/repository"
)

// balanceServiceImpl реализует логику баланса.
type balanceServiceImpl struct {
	balanceRepo *repository.BalanceRepository
}

// NewBalanceService создает новый BalanceService.
func NewBalanceService(balanceRepo *repository.BalanceRepository, _ *repository.OrderRepository) BalanceService {
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

	balance, err := s.balanceRepo.GetBalance(ctx, userID)
	if err != nil {
		return err
	}
	if balance.Current < sum {
		return model.ErrInsufficientFunds
	}

	return s.balanceRepo.Withdraw(ctx, userID, orderNumber, sum)
}

// GetWithdrawals возвращает историю списаний.
func (s *balanceServiceImpl) GetWithdrawals(ctx context.Context, userID int) ([]model.Withdrawal, error) {
	return s.balanceRepo.GetWithdrawals(ctx, userID)
}
