package model

import (
	"time"

	"github.com/google/uuid"
)

// Preferences holds a user's dietary settings. It lives in the domain model
// package and is the type returned by the repository layer. PreferencesDTO in
// dto.go is the serialised view returned to API consumers.
type Preferences struct {
	UserID              uuid.UUID `json:"user_id"`
	DislikedIngredients []string  `json:"disliked_ingredients"`
	MaxPrepMinutes      int       `json:"max_prep_minutes"`
	PreferredCuisines   []string  `json:"preferred_cuisines"`
	UpdatedAt           time.Time `json:"updated_at"`
}