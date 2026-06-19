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
	db *sql.DB
}

// NewPreferencesRepository returns a new PreferencesRepository backed by the provided pool.
func NewPreferencesRepository(db *sql.DB) *PreferencesRepository {
	return &PreferencesRepository{db: db}
}

// GetByUserID returns the user's preferences, creating defaults if missing.
func (r *PreferencesRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (model.Preferences, error) {
	var p model.Preferences
	var dislikedBytes, cuisinesBytes []byte
	var userIDStr string
	err := r.db.QueryRowContext(ctx, `
		SELECT user_id, disliked_ingredients, max_prep_minutes, preferred_cuisines, updated_at
		FROM user_preferences
		WHERE user_id = ?
	`, userID.String()).Scan(&userIDStr, &dislikedBytes, &p.MaxPrepMinutes, &cuisinesBytes, &p.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return r.createDefaults(ctx, userID)
	}
	if err != nil {
		return model.Preferences{}, err
	}

	p.UserID = uuid.MustParse(userIDStr)
	if err := json.Unmarshal(dislikedBytes, &p.DislikedIngredients); err != nil {
		return model.Preferences{}, err
	}
	if err := json.Unmarshal(cuisinesBytes, &p.PreferredCuisines); err != nil {
		return model.Preferences{}, err
	}
	return p, nil
}

// Update replaces the user's preferences.
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
	return r.GetByUserID(ctx, userID)
}

func (r *PreferencesRepository) createDefaults(ctx context.Context, userID uuid.UUID) (model.Preferences, error) {
	p := model.Preferences{
		UserID:              userID,
		DislikedIngredients: []string{},
		MaxPrepMinutes:     30,
		PreferredCuisines:  []string{},
	}
	dislikedBytes, _ := json.Marshal([]string{})
	cuisinesBytes, _ := json.Marshal([]string{})

	_, err := r.db.ExecContext(ctx, `
		INSERT IGNORE INTO user_preferences (user_id, disliked_ingredients, max_prep_minutes, preferred_cuisines)
		VALUES (?, ?, ?, ?)
	`, userID.String(), string(dislikedBytes), 30, string(cuisinesBytes))
	if err != nil {
		return model.Preferences{}, err
	}
	return p, nil
}