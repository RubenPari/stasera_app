package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/stasera/stasera-api/internal/model"
)

// StapleRepository manages persistence for user staples.
type StapleRepository struct {
	db DBTX
}

// NewStapleRepository returns a new StapleRepository backed by the provided pool.
func NewStapleRepository(db *sql.DB) *StapleRepository {
	return &StapleRepository{db: db}
}

// WithTx returns a copy of the repository bound to the provided transaction.
func (r *StapleRepository) WithTx(tx DBTX) *StapleRepository {
	return &StapleRepository{db: tx}
}

// defaultStaples is the list of ingredients seeded for every new user.
var defaultStaples = []string{
	"pasta",
	"riso",
	"olio extravergine d'oliva",
	"sale",
	"pepe",
	"aglio",
	"tonno in scatoletta",
	"pomodori pelati in scatola",
	"uova",
	"pane in cassetta",
	"aceto",
	"dado da brodo",
}

// SeedDefaults inserts the default staple list for the given user, ignoring duplicates.
func (r *StapleRepository) SeedDefaults(ctx context.Context, userID uuid.UUID) error {
	return seedStaples(ctx, r.db, userID, defaultStaples)
}

// seedStaples inserts the given staple names for the user in a single multi-value
// INSERT IGNORE statement. Shared by SeedDefaults and the atomic user-creation path.
func seedStaples(ctx context.Context, db DBTX, userID uuid.UUID, names []string) error {
	if len(names) == 0 {
		return nil
	}
	placeholder := "(?, ?, ?, TRUE)"
	placeholders := make([]string, len(names))
	args := make([]any, 0, len(names)*3)
	for i, name := range names {
		placeholders[i] = placeholder
		args = append(args, uuid.NewString(), userID.String(), name)
	}
	query := "INSERT IGNORE INTO staples (id, user_id, name, is_active) VALUES " + strings.Join(placeholders, ",")
	_, err := db.ExecContext(ctx, query, args...)
	return err
}

// GetActiveByUserID returns all active staples for a user.
func (r *StapleRepository) GetActiveByUserID(ctx context.Context, userID uuid.UUID) ([]model.Staple, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, name, is_active
		FROM staples
		WHERE user_id = ? AND is_active = TRUE
		ORDER BY name
	`, userID.String())
	if err != nil {
		return nil, err
	}
	return Collect(rows, func(s Scanner) (model.Staple, error) {
		st, err := scanStaple(s)
		if err != nil {
			return model.Staple{}, err
		}
		return *st, nil
	})
}

// GetAllByUserID returns all staples for a user, active and inactive.
func (r *StapleRepository) GetAllByUserID(ctx context.Context, userID uuid.UUID) ([]model.Staple, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, name, is_active
		FROM staples
		WHERE user_id = ?
		ORDER BY is_active DESC, name
	`, userID.String())
	if err != nil {
		return nil, err
	}
	return Collect(rows, func(s Scanner) (model.Staple, error) {
		st, err := scanStaple(s)
		if err != nil {
			return model.Staple{}, err
		}
		return *st, nil
	})
}

// Create adds a new staple for the user, ignoring duplicates by name.
// If the staple already exists (even inactive), it is reactivated and returned.
func (r *StapleRepository) Create(ctx context.Context, userID uuid.UUID, name string) (model.Staple, error) {
	id := uuid.NewString()
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO staples (id, user_id, name)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE is_active = TRUE
	`, id, userID.String(), name)
	if err != nil {
		return model.Staple{}, err
	}

	var s model.Staple
	var idStr, userIDStr string
	if err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, name, is_active
		FROM staples
		WHERE user_id = ? AND name = ?
	`, userID.String(), name).Scan(&idStr, &userIDStr, &s.Name, &s.IsActive); err != nil {
		return model.Staple{}, err
	}
	parsedID, err := uuid.Parse(idStr)
	if err != nil {
		return model.Staple{}, err
	}
	parsedUserID, err := uuid.Parse(userIDStr)
	if err != nil {
		return model.Staple{}, err
	}
	s.ID = parsedID
	s.UserID = parsedUserID
	return s, nil
}

// UpdateActive changes the is_active flag of a staple owned by the user.
// Returns nil, nil when the staple does not exist or does not belong to the user.
func (r *StapleRepository) UpdateActive(ctx context.Context, userID, id uuid.UUID, isActive bool) (*model.Staple, error) {
	_, err := r.db.ExecContext(ctx, `
		UPDATE staples
		SET is_active = ?
		WHERE id = ? AND user_id = ?
	`, isActive, id.String(), userID.String())
	if err != nil {
		return nil, err
	}

	row := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, name, is_active
		FROM staples
		WHERE id = ? AND user_id = ?
	`, id.String(), userID.String())
	s, err := scanStaple(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return s, nil
}

// DeleteByID removes a staple owned by the user. Returns false when nothing was deleted.
func (r *StapleRepository) DeleteByID(ctx context.Context, userID, id uuid.UUID) (bool, error) {
	res, err := r.db.ExecContext(ctx, `
		DELETE FROM staples
		WHERE id = ? AND user_id = ?
	`, id.String(), userID.String())
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// scanStaple reads a staple row using the shared Scanner interface.
func scanStaple(s Scanner) (*model.Staple, error) {
	var st model.Staple
	var idStr, userIDStr string
	if err := s.Scan(&idStr, &userIDStr, &st.Name, &st.IsActive); err != nil {
		return nil, err
	}
	parsedID, err := uuid.Parse(idStr)
	if err != nil {
		return nil, err
	}
	parsedUserID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, err
	}
	st.ID = parsedID
	st.UserID = parsedUserID
	return &st, nil
}