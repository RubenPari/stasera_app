package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// RunMigrations applies all pending SQL migrations located in the migrations/ directory.
// Migration state is tracked in the schema_migrations table.
// Each migration file may contain multiple statements separated by ";" — they are
// split and executed one at a time inside a single transaction.
func RunMigrations(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY
		)
	`); err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	files, err := filepath.Glob("migrations/*.sql")
	if err != nil {
		return fmt.Errorf("list migrations: %w", err)
	}
	sort.Strings(files)

	for _, file := range files {
		version := filepath.Base(file)
		if strings.TrimSpace(version) == "" {
			continue
		}

		var applied bool
		if err := db.QueryRowContext(ctx,
			"SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = ?)",
			version,
		).Scan(&applied); err != nil {
			return fmt.Errorf("check migration %s: %w", version, err)
		}
		if applied {
			continue
		}

		sqlBytes, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", version, err)
		}
		stmts := splitStatements(string(sqlBytes))

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin migration %s: %w", version, err)
		}

		for _, stmt := range stmts {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}
			if _, err := tx.ExecContext(ctx, stmt); err != nil {
				_ = tx.Rollback()
				return fmt.Errorf("execute migration %s: %w", version, err)
			}
		}

		if _, err := tx.ExecContext(ctx,
			"INSERT INTO schema_migrations (version) VALUES (?)",
			version,
		); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("record migration %s: %w", version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", version, err)
		}
	}

	return nil
}

// splitStatements splits a SQL script into individual statements on ";".
// It preserves content inside string literals and comments are stripped
// minimally (single-line "--" comments). Sufficient for migration files
// made of CREATE TABLE statements.
func splitStatements(script string) []string {
	var out []string
	var buf strings.Builder
	inSingle := false
	inDouble := false
	inBacktick := false

	flush := func() {
		s := strings.TrimSpace(buf.String())
		if s != "" {
			out = append(out, s)
		}
		buf.Reset()
	}

	lines := strings.Split(script, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Skip full-line comments starting with --.
		if strings.HasPrefix(trimmed, "--") {
			continue
		}
		for i := 0; i < len(line); i++ {
			c := line[i]
			switch c {
			case '\'':
				if !inDouble && !inBacktick {
					inSingle = !inSingle
				}
			case '"':
				if !inSingle && !inBacktick {
					inDouble = !inDouble
				}
			case '`':
				if !inSingle && !inDouble {
					inBacktick = !inBacktick
				}
			case ';':
				if !inSingle && !inDouble && !inBacktick {
					buf.WriteString(line[:i+1])
					flush()
					goto nextLine
				}
			}
		}
		buf.WriteString(line)
		buf.WriteByte('\n')
	nextLine:
	}
	flush()
	return out
}