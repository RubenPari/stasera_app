package repository

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/go-sql-driver/mysql"
)

// isDuplicateEntry reports whether err is a MySQL ER_DUP_ENTRY (1062).
func isDuplicateEntry(err error) bool {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1062
	}
	// Fallback for wrapped errors or text-based checks.
	return err != nil && strings.Contains(err.Error(), "Error 1062")
}

// isNoRows reports whether err is sql.ErrNoRows.
func isNoRows(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}