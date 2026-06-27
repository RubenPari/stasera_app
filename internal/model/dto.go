package model

import (
	"time"

	"github.com/google/uuid"
)

// UserDTO is the public representation of a user, without sensitive fields.
type UserDTO struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	CreatedAt   time.Time `json:"created_at"`
}

// TokenPair contains a signed access token and refresh token.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// AuthResponse is returned by register and login endpoints.
type AuthResponse struct {
	User         UserDTO `json:"user"`
	AccessToken  string  `json:"access_token"`
	RefreshToken string  `json:"refresh_token"`
}

// RecipeIngredientDTO is the public representation of a recipe ingredient.
// It mirrors RecipeIngredient but is isolated from the domain model's JSON tags.
type RecipeIngredientDTO struct {
	Name string `json:"name"`
	Qty  string `json:"qty"`
}

// RecipeStepDTO is the public representation of a recipe step.
type RecipeStepDTO struct {
	Text         string `json:"text"`
	TimerSeconds int    `json:"timer_seconds,omitempty"`
}

// RecipeDTO is the public representation of a recipe.
type RecipeDTO struct {
	ID           uuid.UUID             `json:"id"`
	UserID       uuid.UUID             `json:"user_id"`
	Name         string                `json:"name"`
	PrepMinutes  int                   `json:"prep_minutes"`
	Servings     int                   `json:"servings"`
	Ingredients  []RecipeIngredientDTO `json:"ingredients"`
	Steps        []RecipeStepDTO       `json:"steps"`
	IsRescue     bool                  `json:"is_rescue"`
	TimesCooked  int                   `json:"times_cooked"`
	LastCookedAt *time.Time            `json:"last_cooked_at,omitempty"`
	CreatedAt    time.Time             `json:"created_at"`
}

// MealPlanDTO is the public representation of a meal plan including its days.
type MealPlanDTO struct {
	ID        uuid.UUID        `json:"id"`
	UserID    uuid.UUID        `json:"user_id"`
	WeekStart time.Time        `json:"week_start"`
	Status    string           `json:"status"`
	Days      []MealPlanDayDTO `json:"days"`
	CreatedAt time.Time        `json:"created_at"`
}

// MealPlanDayDTO is the public representation of a meal plan day with optional recipe.
type MealPlanDayDTO struct {
	ID        uuid.UUID  `json:"id"`
	PlanID    uuid.UUID  `json:"plan_id"`
	DayOfWeek int        `json:"day_of_week"`
	RecipeID  uuid.UUID  `json:"recipe_id"`
	Recipe    *RecipeDTO `json:"recipe,omitempty"`
}

// ShoppingItemDTO is the public representation of a shopping list item.
type ShoppingItemDTO struct {
	ID        uuid.UUID `json:"id"`
	ListID    uuid.UUID `json:"list_id"`
	Name      string    `json:"name"`
	Quantity  string    `json:"quantity"`
	Aisle     string    `json:"aisle"`
	IsChecked bool      `json:"is_checked"`
	SortOrder int       `json:"sort_order"`
}

// ShoppingListDTO is the public representation of a shopping list with its items.
type ShoppingListDTO struct {
	ID          uuid.UUID         `json:"id"`
	UserID      uuid.UUID         `json:"user_id"`
	PlanID      *uuid.UUID        `json:"plan_id,omitempty"`
	Items       []ShoppingItemDTO `json:"items"`
	CreatedAt   time.Time         `json:"created_at"`
	CompletedAt *time.Time        `json:"completed_at,omitempty"`
}

// StapleDTO is the public representation of a user staple.
type StapleDTO struct {
	ID       uuid.UUID `json:"id"`
	UserID   uuid.UUID `json:"user_id"`
	Name     string    `json:"name"`
	IsActive bool      `json:"is_active"`
}

// PreferencesDTO is the public representation of user dietary preferences.
// Slice fields are always serialized as arrays (never null).
type PreferencesDTO struct {
	UserID              uuid.UUID `json:"user_id"`
	DislikedIngredients []string  `json:"disliked_ingredients"`
	MaxPrepMinutes      int       `json:"max_prep_minutes"`
	PreferredCuisines   []string  `json:"preferred_cuisines"`
	UpdatedAt           time.Time `json:"updated_at"`
}