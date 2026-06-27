package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/stasera/stasera-api/internal/ai"
	"github.com/stasera/stasera-api/internal/model"
)

// RescueService generates emergency meals from the user's active staples and
// persists them as rescue recipes. It replaces the former AIHandler god-function.
type RescueService struct {
	ai      AIGateway
	recipes RescueRecipeStore
	staples StapleStore
}

// RescueRecipeStore is the recipe persistence surface needed by RescueService.
type RescueRecipeStore interface {
	Create(ctx context.Context, userID uuid.UUID, name string, prepMinutes, servings int, ingredients []model.RecipeIngredient, steps []model.RecipeStep, isRescue bool) (model.Recipe, error)
}

// NewRescueService returns a new RescueService.
func NewRescueService(ai AIGateway, recipes RescueRecipeStore, staples StapleStore) *RescueService {
	return &RescueService{ai: ai, recipes: recipes, staples: staples}
}

// Generate asks the AI for 3 emergency meals using the user's active staples and
// persists each as a rescue recipe. Returns the created recipes.
func (s *RescueService) Generate(ctx context.Context, userID uuid.UUID) ([]model.Recipe, error) {
	activeStaples, err := s.staples.GetActiveByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("load staples: %w", err)
	}

	names := make([]string, len(activeStaples))
	for i, st := range activeStaples {
		names[i] = st.Name
	}

	rawRecipes, err := s.ai.GenerateRescueMeals(ctx, ai.RescueInput{Staples: names})
	if err != nil {
		return nil, fmt.Errorf("AI generation: %w", ErrAIUnavailable)
	}

	recipes := make([]model.Recipe, 0, len(rawRecipes))
	for _, raw := range rawRecipes {
		recipe, err := s.recipes.Create(ctx, userID, raw.Name, raw.PrepMinutes, 1, raw.ToIngredients(), raw.ToSteps(), true)
		if err != nil {
			return nil, fmt.Errorf("save rescue recipe: %w", err)
		}
		recipes = append(recipes, recipe)
	}
	return recipes, nil
}