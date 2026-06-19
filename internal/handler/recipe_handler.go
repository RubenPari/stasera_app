package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/stasera/stasera-api/internal/middleware"
	"github.com/stasera/stasera-api/internal/model"
	"github.com/stasera/stasera-api/internal/repository"
)

// RecipeHandler exposes recipe CRUD endpoints for the authenticated user.
type RecipeHandler struct {
	repo *repository.RecipeRepository
}

// NewRecipeHandler returns a new RecipeHandler.
func NewRecipeHandler(repo *repository.RecipeRepository) *RecipeHandler {
	return &RecipeHandler{repo: repo}
}

// List returns all cached recipes for the user, optionally filtered by is_rescue.
func (h *RecipeHandler) List(c echo.Context) error {
	uid, err := middleware.GetUserID(c)
	if err != nil {
		return middleware.Unauthorized(c, "missing or invalid token")
	}

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
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to load recipes"})
	}

	dtos := make([]model.RecipeDTO, len(recipes))
	for i, r := range recipes {
		dtos[i] = toRecipeDTO(r)
	}
	return c.JSON(http.StatusOK, dtos)
}

// Get returns a single recipe owned by the user.
func (h *RecipeHandler) Get(c echo.Context) error {
	uid, err := middleware.GetUserID(c)
	if err != nil {
		return middleware.Unauthorized(c, "missing or invalid token")
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid recipe id"})
	}

	recipe, err := h.repo.GetByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to load recipe"})
	}
	if recipe == nil || recipe.UserID != uid {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "recipe not found"})
	}

	return c.JSON(http.StatusOK, toRecipeDTO(*recipe))
}

// MarkCooked increments the times_cooked counter and sets last_cooked_at to today.
func (h *RecipeHandler) MarkCooked(c echo.Context) error {
	uid, err := middleware.GetUserID(c)
	if err != nil {
		return middleware.Unauthorized(c, "missing or invalid token")
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid recipe id"})
	}

	existing, err := h.repo.GetByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to load recipe"})
	}
	if existing == nil || existing.UserID != uid {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "recipe not found"})
	}

	recipe, err := h.repo.MarkCooked(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to mark recipe as cooked"})
	}
	return c.JSON(http.StatusOK, toRecipeDTO(recipe))
}

// Delete removes a recipe that is not referenced by any active meal plan.
func (h *RecipeHandler) Delete(c echo.Context) error {
	uid, err := middleware.GetUserID(c)
	if err != nil {
		return middleware.Unauthorized(c, "missing or invalid token")
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid recipe id"})
	}

	existing, err := h.repo.GetByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to load recipe"})
	}
	if existing == nil || existing.UserID != uid {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "recipe not found"})
	}

	if err := h.repo.Delete(c.Request().Context(), id); err != nil {
		if err.Error() == repository.ErrRecipeInUse.Error() {
			return c.JSON(http.StatusConflict, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}