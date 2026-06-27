package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/stasera/stasera-api/internal/model"
	"github.com/stasera/stasera-api/internal/repository"
)

// RecipeHandler exposes recipe CRUD endpoints for the authenticated user.
type RecipeHandler struct {
	repo RecipeStore
}

// NewRecipeHandler returns a new RecipeHandler.
func NewRecipeHandler(repo RecipeStore) *RecipeHandler {
	return &RecipeHandler{repo: repo}
}

// List returns all cached recipes for the user, optionally filtered by is_rescue.
func (h *RecipeHandler) List(c echo.Context) error {
	uid := mustUserID(c)

	var isRescue *bool
	if v := c.QueryParam("is_rescue"); v == "true" {
		b := true
		isRescue = &b
	} else if v == "false" {
		b := false
		isRescue = &b
	}

	recipes, err := h.repo.GetByUserID(c.Request().Context(), uid, isRescue)
	if err != nil {
		return respondError(c, err)
	}

	dtos := make([]model.RecipeDTO, len(recipes))
	for i, r := range recipes {
		dtos[i] = toRecipeDTO(r)
	}
	return c.JSON(http.StatusOK, dtos)
}

// Get returns a single recipe owned by the user.
func (h *RecipeHandler) Get(c echo.Context) error {
	uid := mustUserID(c)

	id, err := parsePathUUID(c, "id")
	if err != nil {
		return err
	}

	recipe, err := h.repo.GetByIDForUser(c.Request().Context(), id, uid)
	if err != nil {
		return respondError(c, err)
	}
	if recipe == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "recipe not found"})
	}

	return c.JSON(http.StatusOK, toRecipeDTO(*recipe))
}

// MarkCooked increments the times_cooked counter and sets last_cooked_at to today.
func (h *RecipeHandler) MarkCooked(c echo.Context) error {
	uid := mustUserID(c)

	id, err := parsePathUUID(c, "id")
	if err != nil {
		return err
	}

	recipe, err := h.repo.MarkCookedForUser(c.Request().Context(), id, uid)
	if err != nil {
		return respondError(c, err)
	}
	if recipe == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "recipe not found"})
	}
	return c.JSON(http.StatusOK, toRecipeDTO(*recipe))
}

// Delete removes a recipe that is not referenced by any active meal plan.
func (h *RecipeHandler) Delete(c echo.Context) error {
	uid := mustUserID(c)

	id, err := parsePathUUID(c, "id")
	if err != nil {
		return err
	}

	deleted, err := h.repo.DeleteForUser(c.Request().Context(), id, uid)
	if err != nil {
		if errors.Is(err, repository.ErrRecipeInUse) {
			return c.JSON(http.StatusConflict, map[string]string{"error": "recipe in use"})
		}
		return respondError(c, err)
	}
	if !deleted {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "recipe not found"})
	}
	return c.NoContent(http.StatusNoContent)
}

// toRecipeDTO converts a domain recipe to its API representation.
// Ingredient and step slices are always non-nil (serialised as [] not null).
func toRecipeDTO(r model.Recipe) model.RecipeDTO {
	ings := make([]model.RecipeIngredientDTO, len(r.Ingredients))
	for i, ing := range r.Ingredients {
		ings[i] = model.RecipeIngredientDTO{Name: ing.Name, Qty: ing.Qty}
	}
	steps := make([]model.RecipeStepDTO, len(r.Steps))
	for i, st := range r.Steps {
		steps[i] = model.RecipeStepDTO{Text: st.Text, TimerSeconds: st.TimerSeconds}
	}
	return model.RecipeDTO{
		ID:           r.ID,
		UserID:       r.UserID,
		Name:         r.Name,
		PrepMinutes:  r.PrepMinutes,
		Servings:     r.Servings,
		Ingredients:  ings,
		Steps:        steps,
		IsRescue:     r.IsRescue,
		TimesCooked:  r.TimesCooked,
		LastCookedAt: r.LastCookedAt,
		CreatedAt:    r.CreatedAt,
	}
}