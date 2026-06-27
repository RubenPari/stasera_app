package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

// RunMigrations applies all pending SQL migrations embedded in the binary.
// Migration state is tracked in the schema_migrations table.
// A MySQL advisory lock prevents concurrent executions on multi-replica deploys.
func RunMigrations(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`); err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	// Advisory lock ensures only one replica runs migrations at a time.
	var locked int
	if err := db.QueryRowContext(ctx, "SELECT GET_LOCK('stasera_migrations', 10)").Scan(&locked); err != nil {
		return fmt.Errorf("acquire migration lock: %w", err)
	}
	if locked != 1 {
		return fmt.Errorf("could not acquire migration lock within 10 seconds")
	}
	defer db.ExecContext(ctx, "SELECT RELEASE_LOCK('stasera_migrations')") //nolint:errcheck

	entries, err := fs.ReadDir(migrationFS, "migrations")
	if err != nil {
		return fmt.Errorf("list migrations: %w", err)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		version := entry.Name()

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

		sqlBytes, err := migrationFS.ReadFile("migrations/" + version)
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
// It handles string literals (single quotes, double quotes, backticks) and strips
// full-line "--" comments. Sufficient for CREATE TABLE migration files.
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

	for _, line := range strings.Split(script, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "--") {
			continue
		}
		flushed := false
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
					buf.WriteByte(c)
					flush()
					// Append remainder of the line after the semicolon.
					if rest := strings.TrimSpace(line[i+1:]); rest != "" {
						buf.WriteString(rest)
						buf.WriteByte('\n')
					}
					flushed = true
					goto nextLine
				}
			}
			buf.WriteByte(c)
		}
		if !flushed {
			buf.WriteByte('\n')
		}
	nextLine:
	}
	flush()
	return out
}