package repository

import (
	"database/sql"

	_ "github.com/lib/pq"
)

// NewDB создает новое подключение к PostgreSQL.
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
