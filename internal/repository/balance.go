package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/tempizhere/gofemarket/internal/model"
	"go.uber.org/zap"
)

// BalanceRepository управляет данными баланса в БД.
type BalanceRepository struct {
	db *sql.DB
}

// NewBalanceRepository создает новый BalanceRepository.
func NewBalanceRepository(db *sql.DB) *BalanceRepository {
	return &BalanceRepository{db: db}
}

// GetBalance возвращает текущий баланс.
func (r *BalanceRepository) GetBalance(ctx context.Context, userID int) (model.Balance, error) {
	var balance model.Balance
	err := r.db.QueryRowContext(ctx,
		"SELECT current, withdrawn FROM balances WHERE user_id = $1",
		userID,
	).Scan(&balance.Current, &balance.Withdrawn)
	if err == sql.ErrNoRows {
		return model.Balance{}, nil
	}
	return balance, err
}

// Withdraw выполняет списание баллов.
func (r *BalanceRepository) Withdraw(ctx context.Context, userID int, orderNumber string, sum float64) error {
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
		"UPDATE balances SET current = current - $1, withdrawn = withdrawn + $1 WHERE user_id = $2",
		sum, userID,
	)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx,
		"INSERT INTO withdrawals (user_id, order_number, sum, processed_at) VALUES ($1, $2, $3, $4)",
		userID, orderNumber, sum, time.Now(),
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// GetWithdrawals возвращает историю списаний.
func (r *BalanceRepository) GetWithdrawals(ctx context.Context, userID int) ([]model.Withdrawal, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT order_number, sum, processed_at FROM withdrawals WHERE user_id = $1 ORDER BY processed_at DESC",
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

	var withdrawals []model.Withdrawal
	for rows.Next() {
		var w model.Withdrawal
		if err := rows.Scan(&w.Order, &w.Sum, &w.ProcessedAt); err != nil {
			return nil, err
		}
		withdrawals = append(withdrawals, w)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return withdrawals, nil
}
