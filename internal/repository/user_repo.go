package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/stasera/stasera-api/internal/model"
)

// UserRepository manages persistence for user accounts.
// pool is the underlying *sql.DB used to begin transactions for atomic multi-step
// operations (e.g. CreateWithDefaultStaples); db is the execution handle and may be a tx.
type UserRepository struct {
	db   DBTX
	pool *sql.DB
}

// NewUserRepository returns a new UserRepository backed by the provided pool.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db, pool: db}
}

// WithTx returns a copy of the repository bound to the provided transaction.
func (r *UserRepository) WithTx(tx DBTX) *UserRepository {
	return &UserRepository{db: tx, pool: nil}
}

// Create inserts a new user and returns the created record.
func (r *UserRepository) Create(ctx context.Context, email, passwordHash, displayName string) (model.User, error) {
	id := uuid.NewString()
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash, display_name)
		VALUES (?, ?, ?, ?)
	`, id, email, passwordHash, displayName)
	if err != nil {
		if isDuplicateEntry(err) {
			return model.User{}, ErrDuplicateEmail
		}
		return model.User{}, err
	}
	uid, err := uuid.Parse(id)
	if err != nil {
		return model.User{}, err
	}
	return model.User{
		ID:           uid,
		Email:        email,
		PasswordHash: passwordHash,
		DisplayName:  displayName,
		CreatedAt:    time.Now(),
	}, nil
}

// CreateWithDefaultStaples atomically creates a user and seeds the default
// staples within a single transaction, so a seed failure rolls back the user
// and never leaves an orphan account.
func (r *UserRepository) CreateWithDefaultStaples(ctx context.Context, email, passwordHash, displayName string) (model.User, error) {
	if r.pool == nil {
		return model.User{}, fmt.Errorf("user repository has no pool to begin a transaction")
	}
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return model.User{}, fmt.Errorf("begin tx: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	id := uuid.NewString()
	_, err = tx.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash, display_name)
		VALUES (?, ?, ?, ?)
	`, id, email, passwordHash, displayName)
	if err != nil {
		if isDuplicateEntry(err) {
			return model.User{}, ErrDuplicateEmail
		}
		return model.User{}, err
	}

	uid, err := uuid.Parse(id)
	if err != nil {
		return model.User{}, err
	}
	if err := seedStaples(ctx, tx, uid, defaultStaples); err != nil {
		return model.User{}, err
	}

	if err := tx.Commit(); err != nil {
		return model.User{}, err
	}
	committed = true

	return model.User{
		ID:           uid,
		Email:        email,
		PasswordHash: passwordHash,
		DisplayName:  displayName,
		CreatedAt:    time.Now(),
	}, nil
}

// FindByEmail looks up a user by email address.
// Returns nil, nil when no user is found.
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, email, password_hash, display_name, created_at
		FROM users
		WHERE email = ?
	`, email)
	u, err := scanUser(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

// GetByID looks up a user by primary key.
// Returns nil, nil when no user is found.
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, email, password_hash, display_name, created_at
		FROM users
		WHERE id = ?
	`, id.String())
	u, err := scanUser(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

// UpdateProfile updates the user's display_name and returns the refreshed record.
func (r *UserRepository) UpdateProfile(ctx context.Context, id uuid.UUID, displayName string) (model.User, error) {
	if _, err := r.db.ExecContext(ctx, `
		UPDATE users
		SET display_name = ?
		WHERE id = ?
	`, displayName, id.String()); err != nil {
		return model.User{}, err
	}
	u, err := r.GetByID(ctx, id)
	if err != nil {
		return model.User{}, err
	}
	if u == nil {
		return model.User{}, sql.ErrNoRows
	}
	return *u, nil
}

// UpdatePasswordHash replaces the stored password hash for a user.
func (r *UserRepository) UpdatePasswordHash(ctx context.Context, id uuid.UUID, hash string) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE users
		SET password_hash = ?
		WHERE id = ?
	`, hash, id.String())
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}


// scanUser reads a user row from a single-row scanner.
func scanUser(row interface {
	Scan(dest ...interface{}) error
}) (*model.User, error) {
	var u model.User
	var idStr string
	if err := row.Scan(&idStr, &u.Email, &u.PasswordHash, &u.DisplayName, &u.CreatedAt); err != nil {
		return nil, err
	}
	parsed, err := uuid.Parse(idStr)
	if err != nil {
		return nil, err
	}
	u.ID = parsed
	return &u, nil
}
