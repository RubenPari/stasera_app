package repository

import "database/sql"

// Scanner is the common interface satisfied by both *sql.Row and *sql.Rows.
type Scanner interface {
	Scan(dest ...any) error
}

// Collect iterates rows, applies scan to each, and returns a slice.
// It closes rows on return (including on error) and propagates rows.Err().
func Collect[T any](rows *sql.Rows, scan func(Scanner) (T, error)) ([]T, error) {
	defer rows.Close()
	var out []T
	for rows.Next() {
		v, err := scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}