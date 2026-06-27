package repository

import (
	"errors"
	"strings"

	"github.com/go-sql-driver/mysql"
)

// ErrDuplicateEmail is a sentinel returned by UserRepository.Create when the
// email is already registered. The handler layer maps it to HTTP 409.
var ErrDuplicateEmail = errors.New("email already registered")

// ErrRecipeInUse is returned when trying to delete a recipe referenced by a
// meal plan. The handler layer maps it to HTTP 409.
var ErrRecipeInUse = errors.New("recipe is used in a meal plan")

// isDuplicateEntry reports whether err is a MySQL ER_DUP_ENTRY (1062).
func isDuplicateEntry(err error) bool {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1062
	}
	// Fallback for wrapped errors or text-based checks.
	return err != nil && strings.Contains(err.Error(), "Error 1062")
}