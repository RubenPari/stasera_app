package model

import (
	"time"

	"github.com/google/uuid"
)

// RecipeIngredient represents a single ingredient inside a recipe.
// The db: tag is intentionally absent — scanners in recipe_repo.go use
// manual field-pointer scanning, not struct-scanning.
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
	ID           uuid.UUID          `json:"id"`
	UserID       uuid.UUID          `json:"user_id"`
	Name         string             `json:"name"`
	PrepMinutes  int                `json:"prep_minutes"`
	Servings     int                `json:"servings"`
	Ingredients  []RecipeIngredient `json:"ingredients"`
	Steps        []RecipeStep       `json:"steps"`
	IsRescue     bool               `json:"is_rescue"`
	TimesCooked  int                `json:"times_cooked"`
	LastCookedAt *time.Time         `json:"last_cooked_at,omitempty"`
	CreatedAt    time.Time          `json:"created_at"`
}