package repository

import (
	"context"
	"database/sql"

	"github.com/tempizhere/gofemarket/internal/model"
	"go.uber.org/zap"
)

// UserRepository управляет данными пользователей в БД.
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository создает новый UserRepository.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUserWithBalance создает пользователя и запись баланса в транзакции.
func (r *UserRepository) CreateUserWithBalance(ctx context.Context, login, password string) (int, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			zap.L().Error("failed to rollback transaction", zap.Error(err))
		}
	}()

	var userID int
	err = tx.QueryRowContext(ctx,
		"INSERT INTO users (login, password) VALUES ($1, $2) RETURNING id",
		login, password,
	).Scan(&userID)
	if err != nil {
		if err.Error() == `pq: duplicate key value violates unique constraint "users_login_key"` {
			return 0, model.ErrLoginTaken
		}
		return 0, err
	}

	_, err = tx.ExecContext(ctx,
		"INSERT INTO balances (user_id, current, withdrawn) VALUES ($1, 0, 0) ON CONFLICT DO NOTHING",
		userID,
	)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return userID, nil
}

// GetUserByLogin получает пользователя по логину.
func (r *UserRepository) GetUserByLogin(ctx context.Context, login string) (model.User, error) {
	var user model.User
	err := r.db.QueryRowContext(ctx,
		"SELECT id, login, password FROM users WHERE login = $1",
		login,
	).Scan(&user.ID, &user.Login, &user.Password)
	if err == sql.ErrNoRows {
		return model.User{}, model.ErrInvalidCredentials
	}
	return user, err
}
