package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/stasera/stasera-api/internal/model"
)

// PreferencesRepository manages persistence for user preferences.
type PreferencesRepository struct {
	db DBTX
}

// NewPreferencesRepository returns a new PreferencesRepository backed by the provided pool.
func NewPreferencesRepository(db *sql.DB) *PreferencesRepository {
	return &PreferencesRepository{db: db}
}

// WithTx returns a copy of the repository bound to the provided transaction.
func (r *PreferencesRepository) WithTx(tx DBTX) *PreferencesRepository {
	return &PreferencesRepository{db: tx}
}

// GetByUserID returns the user's preferences.
// Returns nil, nil when no preferences row exists for the user (the handler maps
// this to 404; callers needing defaults must apply them explicitly).
func (r *PreferencesRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*model.Preferences, error) {
	var p model.Preferences
	var dislikedBytes, cuisinesBytes []byte
	var userIDStr string
	err := r.db.QueryRowContext(ctx, `
		SELECT user_id, disliked_ingredients, max_prep_minutes, preferred_cuisines, updated_at
		FROM user_preferences
		WHERE user_id = ?
	`, userID.String()).Scan(&userIDStr, &dislikedBytes, &p.MaxPrepMinutes, &cuisinesBytes, &p.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, err
	}
	p.UserID = uid
	if err := json.Unmarshal(dislikedBytes, &p.DislikedIngredients); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(cuisinesBytes, &p.PreferredCuisines); err != nil {
		return nil, err
	}
	return &p, nil
}

// Update replaces the user's preferences (upsert) and returns the updated record.
func (r *PreferencesRepository) Update(ctx context.Context, userID uuid.UUID, p model.Preferences) (model.Preferences, error) {
	dislikedBytes, err := json.Marshal(p.DislikedIngredients)
	if err != nil {
		return model.Preferences{}, err
	}
	cuisinesBytes, err := json.Marshal(p.PreferredCuisines)
	if err != nil {
		return model.Preferences{}, err
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO user_preferences (user_id, disliked_ingredients, max_prep_minutes, preferred_cuisines)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			disliked_ingredients = VALUES(disliked_ingredients),
			max_prep_minutes = VALUES(max_prep_minutes),
			preferred_cuisines = VALUES(preferred_cuisines),
			updated_at = NOW()
	`, userID.String(), string(dislikedBytes), p.MaxPrepMinutes, string(cuisinesBytes))
	if err != nil {
		return model.Preferences{}, err
	}
	updated, err := r.GetByUserID(ctx, userID)
	if err != nil {
		return model.Preferences{}, err
	}
	if updated == nil {
		return model.Preferences{}, sql.ErrNoRows
	}
	return *updated, nil
}