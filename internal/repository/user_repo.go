package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/stasera/stasera-api/internal/model"
)

// UserRepository manages persistence for user accounts.
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository returns a new UserRepository backed by the provided pool.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
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
	u, err := r.FindByID(ctx, uuid.MustParse(id))
	if err != nil {
		return model.User{}, err
	}
	return *u, nil
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

// FindByID looks up a user by primary key.
// Returns nil, nil when no user is found.
func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
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
	u, err := r.FindByID(ctx, id)
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

// ErrDuplicateEmail is a sentinel error mapped by the handler to HTTP 409.
var ErrDuplicateEmail = errors.New("email already registered")

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
