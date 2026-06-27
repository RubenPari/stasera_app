package repository

import (
	"context"
	"database/sql"
)

// DBTX is the minimal database surface implemented by both *sql.DB and *sql.Tx.
// Repositories depend on this interface so service-layer code can run a set of
// operations inside a single transaction by passing a *sql.Tx via WithTx.
type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}