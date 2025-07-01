package repository

import (
	"context"
	"database/sql"

	"github.com/tempizhere/gofemarket/internal/model"
)

// UserRepository управляет данными пользователей в БД.
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository создает новый UserRepository.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser создает нового пользователя.
func (r *UserRepository) CreateUser(ctx context.Context, login, password string) (int, error) {
	var userID int
	err := r.db.QueryRowContext(ctx,
		"INSERT INTO users (login, password) VALUES ($1, $2) RETURNING id",
		login, password,
	).Scan(&userID)
	if err != nil {
		if err.Error() == `pq: duplicate key value violates unique constraint "users_login_key"` {
			return 0, model.ErrLoginTaken
		}
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
