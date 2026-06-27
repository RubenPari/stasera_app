package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/stasera/stasera-api/internal/ai"
	"github.com/stasera/stasera-api/internal/model"
)

// AIGateway is the consumer-side interface for the AI provider, used by services
// that orchestrate generation. It mirrors the methods of *ai.Gateway so a fake
// can be substituted in tests.
type AIGateway interface {
	GenerateMealPlan(ctx context.Context, input ai.MealPlanInput) ([]ai.RawRecipe, error)
	GenerateRescueMeals(ctx context.Context, input ai.RescueInput) ([]ai.RawRecipe, error)
	GenerateSingleRecipe(ctx context.Context, input ai.MealPlanInput, dayOfWeek int) (ai.RawRecipe, error)
}

// MealPlanStore is the persistence surface used by MealPlanService.
type MealPlanStore interface {
	Create(ctx context.Context, userID uuid.UUID, weekStart time.Time) (model.MealPlan, error)
	GetCurrent(ctx context.Context, userID uuid.UUID) (*model.MealPlan, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.MealPlan, error)
	AddDay(ctx context.Context, planID uuid.UUID, dayOfWeek int, recipeID uuid.UUID) (model.MealPlanDay, error)
	GetDays(ctx context.Context, planID uuid.UUID) ([]model.MealPlanDay, error)
	UpdateDayRecipe(ctx context.Context, planID uuid.UUID, dayOfWeek int, recipeID uuid.UUID) (model.MealPlanDay, error)
	GetTodayRecipeID(ctx context.Context, userID uuid.UUID) (*uuid.UUID, error)
	ArchiveOldPlans(ctx context.Context, userID uuid.UUID, weekStart time.Time) error
	PlanExistsForWeek(ctx context.Context, userID uuid.UUID, weekStart time.Time) (bool, error)
}

// RecipeStore is the recipe persistence surface used by services.
type RecipeStore interface {
	Create(ctx context.Context, userID uuid.UUID, name string, prepMinutes, servings int, ingredients []model.RecipeIngredient, steps []model.RecipeStep, isRescue bool) (model.Recipe, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Recipe, error)
	GetByPlanID(ctx context.Context, planID uuid.UUID) ([]model.Recipe, error)
	FindRecentNames(ctx context.Context, userID uuid.UUID) ([]string, error)
	GetIngredientsByPlanID(ctx context.Context, planID uuid.UUID) ([]model.RecipeIngredient, error)
}

// PreferencesStore is the preferences persistence surface used by services.
// Returns nil, nil when the user has no preferences row; callers needing defaults
// must apply them explicitly.
type PreferencesStore interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (*model.Preferences, error)
}

// StapleStore is the staple persistence surface used by services.
type StapleStore interface {
	GetActiveByUserID(ctx context.Context, userID uuid.UUID) ([]model.Staple, error)
}

// ShoppingListStore is the shopping-list persistence surface used by ShoppingListService.
type ShoppingListStore interface {
	GetCurrent(ctx context.Context, userID uuid.UUID) (*model.ShoppingList, error)
	DeleteOpenByUserID(ctx context.Context, userID uuid.UUID) error
	CreateWithItems(ctx context.Context, userID uuid.UUID, planID *uuid.UUID, items []model.ShoppingItem) (model.ShoppingList, error)
	GetItemsByListID(ctx context.Context, listID uuid.UUID) ([]model.ShoppingItem, error)
	UpdateItemChecked(ctx context.Context, userID, itemID uuid.UUID, isChecked bool) (*model.ShoppingItem, error)
	MarkCompleted(ctx context.Context, userID uuid.UUID) (*model.ShoppingList, error)
}