package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/stasera/stasera-api/internal/ai"
	"github.com/stasera/stasera-api/internal/middleware"
	"github.com/stasera/stasera-api/internal/model"
	"github.com/stasera/stasera-api/internal/repository"
)

// AIHandler handles AI-powered endpoints.
type AIHandler struct {
	ai        *ai.Gateway
	recipes   *repository.RecipeRepository
	staples   *repository.StapleRepository
}

// NewAIHandler returns a new AIHandler.
func NewAIHandler(ai *ai.Gateway, recipes *repository.RecipeRepository, staples *repository.StapleRepository) *AIHandler {
	return &AIHandler{
		ai:      ai,
		recipes: recipes,
		staples: staples,
	}
}

// Rescue generates 3 emergency meals using only the user's active staples.
func (h *AIHandler) Rescue(c echo.Context) error {
	uid, err := middleware.GetUserID(c)
	if err != nil {
		return middleware.Unauthorized(c, "missing or invalid token")
	}

	activeStaples, err := h.staples.GetActiveByUserID(c.Request().Context(), uid)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to load staples"})
	}

	names := make([]string, len(activeStaples))
	for i, s := range activeStaples {
		names[i] = s.Name
	}

	rawRecipes, err := h.ai.GenerateRescueMeals(c.Request().Context(), ai.RescueInput{Staples: names})
	if err != nil {
		return c.JSON(http.StatusBadGateway, map[string]string{"error": err.Error()})
	}

	recipes := make([]model.RecipeDTO, 0, len(rawRecipes))
	for _, raw := range rawRecipes {
		recipe, err := h.recipes.Create(
			c.Request().Context(),
			uid,
			raw.Name,
			raw.PrepMinutes,
			1,
			rawToIngredients(raw.Ingredients),
			rawToSteps(raw.Steps),
			true,
		)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to save rescue recipe"})
		}
		recipes = append(recipes, toRecipeDTO(recipe))
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"recipes": recipes})
}

func rawToIngredients(raw []map[string]string) []model.RecipeIngredient {
	out := make([]model.RecipeIngredient, 0, len(raw))
	for _, m := range raw {
		out = append(out, model.RecipeIngredient{
			Name: m["name"],
			Qty:  m["qty"],
		})
	}
	return out
}

func rawToSteps(raw []map[string]interface{}) []model.RecipeStep {
	out := make([]model.RecipeStep, 0, len(raw))
	for _, m := range raw {
		text, _ := m["text"].(string)
		var timer int
		if v, ok := m["timer_seconds"].(float64); ok {
			timer = int(v)
		}
		out = append(out, model.RecipeStep{
			Text:         text,
			TimerSeconds: timer,
		})
	}
	return out
}
