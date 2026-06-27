package db

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/stasera/stasera-api/internal/config"
)

// NewPool opens a new MySQL connection pool from the provided DSN and applies
// recommended connection pool limits from the application config.
// The DSN must include parseTime=true (validated in config.Load).
func NewPool(ctx context.Context, cfg *config.Config) (*sql.DB, error) {
	pool, err := sql.Open("mysql", cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	pool.SetMaxOpenConns(cfg.DBMaxOpen)
	pool.SetMaxIdleConns(cfg.DBMaxIdle)
	pool.SetConnMaxLifetime(time.Duration(cfg.DBConnMaxLifetimeSec) * time.Second)

	if err := pool.PingContext(ctx); err != nil {
		_ = pool.Close()
		return nil, err
	}
	return pool, nil
}