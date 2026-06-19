package model

import (
	"time"

	"github.com/google/uuid"
)

// RecipeIngredient represents a single ingredient inside a recipe.
type RecipeIngredient struct {
	Name string `json:"name"`
	Qty  string `json:"qty"`
}

// RecipeStep represents one step in the preparation of a recipe.
type RecipeStep struct {
	Text         string `json:"text"`
	TimerSeconds int    `json:"timer_seconds,omitempty"`
}

// Recipe stores a generated meal cached for a specific user.
type Recipe struct {
	ID           uuid.UUID          `db:"id"             json:"id"`
	UserID       uuid.UUID          `db:"user_id"        json:"user_id"`
	Name         string             `db:"name"           json:"name"`
	PrepMinutes  int                `db:"prep_minutes"   json:"prep_minutes"`
	Servings     int                `db:"servings"       json:"servings"`
	Ingredients  []RecipeIngredient `db:"ingredients"    json:"ingredients"`
	Steps        []RecipeStep       `db:"steps"          json:"steps"`
	IsRescue     bool               `db:"is_rescue"      json:"is_rescue"`
	TimesCooked  int                `db:"times_cooked"   json:"times_cooked"`
	LastCookedAt *time.Time         `db:"last_cooked_at" json:"last_cooked_at,omitempty"`
	CreatedAt    time.Time          `db:"created_at"     json:"created_at"`
}
