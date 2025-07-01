package service

import (
	"context"

	"github.com/tempizhere/gofemarket/internal/model"
)

// UserService определяет интерфейс для работы с пользователями.
type UserService interface {
	Register(ctx context.Context, login, password string) (string, error)
	Login(ctx context.Context, login, password string) (string, error)
	ValidateToken(token string) (int, error)
}

// OrderService определяет интерфейс для работы с заказами.
type OrderService interface {
	UploadOrder(ctx context.Context, userID int, orderNumber string) error
	GetOrders(ctx context.Context, userID int) ([]model.Order, error)
	ProcessOrders(ctx context.Context) error
}

// BalanceService определяет интерфейс для работы с балансом.
type BalanceService interface {
	GetBalance(ctx context.Context, userID int) (model.Balance, error)
	Withdraw(ctx context.Context, userID int, orderNumber string, sum float64) error
	GetWithdrawals(ctx context.Context, userID int) ([]model.Withdrawal, error)
}
