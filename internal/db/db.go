package db

import (
	"context"
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

// NewPool opens a new MySQL connection pool from the provided DSN.
// The DSN must include parseTime=true to map TIMESTAMP/DATETIME to time.Time.
func NewPool(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

// Ping verifies that the pool can reach the database.
func Ping(ctx context.Context, db *sql.DB) error {
	return db.PingContext(ctx)
}

// Close shuts down the connection pool.
func Close(db *sql.DB) {
	_ = db.Close()
}