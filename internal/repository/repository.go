package repository

import (
	"database/sql"

	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

// NewDB создает новое подключение к базе данных.
func NewDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

// InitTables создает необходимые таблицы в базе данных.
func InitTables(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			login TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS orders (
			id SERIAL PRIMARY KEY,
			number TEXT NOT NULL UNIQUE,
			user_id INTEGER REFERENCES users(id),
			status TEXT NOT NULL,
			accrual FLOAT DEFAULT 0,
			uploaded_at TIMESTAMP NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status)`,
		`CREATE TABLE IF NOT EXISTS balances (
			user_id INTEGER PRIMARY KEY REFERENCES users(id),
			current FLOAT DEFAULT 0,
			withdrawn FLOAT DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS withdrawals (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id),
			order_number TEXT NOT NULL,
			sum FLOAT NOT NULL,
			processed_at TIMESTAMP NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_withdrawals_user_id ON withdrawals(user_id)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			zap.L().Error("failed to execute query", zap.String("query", query), zap.Error(err))
			return err
		}
	}
	return nil
}