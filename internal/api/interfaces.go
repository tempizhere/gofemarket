package api

import (
	"context"
	"time"

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

// BalanceRepository определяет интерфейс для работы с балансом в БД.
type BalanceRepository interface {
	GetBalance(ctx context.Context, userID int) (model.Balance, error)
	WithdrawWithCheck(ctx context.Context, userID int, orderNumber string, sum float64) error
	GetWithdrawals(ctx context.Context, userID int) ([]model.Withdrawal, error)
}

// OrderRepository определяет интерфейс для работы с заказами в БД.
type OrderRepository interface {
	GetOrderByNumber(ctx context.Context, orderNumber string) (model.Order, error)
	CreateOrder(ctx context.Context, userID int, orderNumber string, status string) error
	GetOrdersByUser(ctx context.Context, userID int) ([]model.Order, error)
	GetOrdersByStatus(ctx context.Context, statuses []string) ([]model.Order, error)
	UpdateOrderAndBalance(ctx context.Context, orderNumber string, status string, accrual float64, userID int) error
}

// UserRepository определяет интерфейс для работы с пользователями в БД.
type UserRepository interface {
	CreateUserWithBalance(ctx context.Context, login, password string) (int, error)
	GetUserByLogin(ctx context.Context, login string) (model.User, error)
}

// AccrualClient определяет интерфейс для взаимодействия с системой начислений.
type AccrualClient interface {
	GetAccrual(ctx context.Context, orderNumber string) (model.Order, time.Duration, error)
}
