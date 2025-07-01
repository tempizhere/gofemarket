package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"
	"github.com/tempizhere/gofemarket/internal/model"
	"go.uber.org/zap"
)

// OrderRepository управляет данными заказов в БД.
type OrderRepository struct {
	db *sql.DB
}

// NewOrderRepository создает новый OrderRepository.
func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// CreateOrder создает новый заказ.
func (r *OrderRepository) CreateOrder(ctx context.Context, userID int, number, status string) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO orders (number, user_id, status, uploaded_at) VALUES ($1, $2, $3, $4)",
		number, userID, status, time.Now(),
	)
	return err
}

// GetOrderByNumber получает заказ по номеру.
func (r *OrderRepository) GetOrderByNumber(ctx context.Context, number string) (model.Order, error) {
	var order model.Order
	err := r.db.QueryRowContext(ctx,
		"SELECT id, number, user_id, status, accrual, uploaded_at FROM orders WHERE number = $1",
		number,
	).Scan(&order.ID, &order.Number, &order.UserID, &order.Status, &order.Accrual, &order.UploadedAt)
	if err == sql.ErrNoRows {
		return model.Order{}, sql.ErrNoRows
	}
	return order, err
}

// GetOrdersByUser возвращает заказы пользователя.
func (r *OrderRepository) GetOrdersByUser(ctx context.Context, userID int) ([]model.Order, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, number, status, accrual, uploaded_at FROM orders WHERE user_id = $1 ORDER BY uploaded_at DESC",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			zap.L().Error("failed to close rows", zap.Error(err))
		}
	}()

	var orders []model.Order
	for rows.Next() {
		var order model.Order
		if err := rows.Scan(&order.ID, &order.Number, &order.Status, &order.Accrual, &order.UploadedAt); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return orders, nil
}

// GetOrdersByStatus возвращает заказы по статусам.
func (r *OrderRepository) GetOrdersByStatus(ctx context.Context, statuses []string) ([]model.Order, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, number, user_id, status, accrual, uploaded_at FROM orders WHERE status = ANY($1)",
		pq.Array(statuses),
	)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			zap.L().Error("failed to close rows", zap.Error(err))
		}
	}()

	var orders []model.Order
	for rows.Next() {
		var order model.Order
		if err := rows.Scan(&order.ID, &order.Number, &order.UserID, &order.Status, &order.Accrual, &order.UploadedAt); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return orders, nil
}

// UpdateOrderAndBalance обновляет заказ и баланс в транзакции.
func (r *OrderRepository) UpdateOrderAndBalance(ctx context.Context, number, status string, accrual float64, userID int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			zap.L().Error("failed to rollback transaction", zap.Error(err))
		}
	}()

	_, err = tx.ExecContext(ctx,
		"UPDATE orders SET status = $1, accrual = $2 WHERE number = $3",
		status, accrual, number,
	)
	if err != nil {
		return err
	}

	if accrual > 0 {
		_, err = tx.ExecContext(ctx,
			"UPDATE balances SET current = current + $1 WHERE user_id = $2",
			accrual, userID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
