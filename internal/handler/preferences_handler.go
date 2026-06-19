package handler

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"

	"github.com/stasera/stasera-api/internal/middleware"
	"github.com/stasera/stasera-api/internal/model"
	"github.com/stasera/stasera-api/internal/repository"
)

// PreferencesHandler exposes endpoints to read and update dietary preferences.
type PreferencesHandler struct {
	repo      *repository.PreferencesRepository
	validator *validator.Validate
}

// NewPreferencesHandler returns a new PreferencesHandler.
func NewPreferencesHandler(repo *repository.PreferencesRepository) *PreferencesHandler {
	return &PreferencesHandler{repo: repo, validator: validator.New()}
}

type updatePreferencesRequest struct {
	DislikedIngredients []string `json:"disliked_ingredients"`
	MaxPrepMinutes      *int      `json:"max_prep_minutes" validate:"required,min=1,max=120"`
	PreferredCuisines   []string `json:"preferred_cuisines"`
}

// Get returns the user's preferences, creating defaults if missing.
func (h *PreferencesHandler) Get(c echo.Context) error {
	uid, err := middleware.GetUserID(c)
	if err != nil {
		return middleware.Unauthorized(c, "missing or invalid token")
	}

	prefs, err := h.repo.GetByUserID(c.Request().Context(), uid)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to load preferences"})
	}
	return c.JSON(http.StatusOK, toPreferencesDTO(prefs))
}

// Update replaces the user's preferences with the provided values.
func (h *PreferencesHandler) Update(c echo.Context) error {
	uid, err := middleware.GetUserID(c)
	if err != nil {
		return middleware.Unauthorized(c, "missing or invalid token")
	}

	var req updatePreferencesRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	if err := h.validator.Struct(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	disliked := req.DislikedIngredients
	if disliked == nil {
		disliked = []string{}
	}
	cuisines := req.PreferredCuisines
	if cuisines == nil {
		cuisines = []string{}
	}

	prefs := model.Preferences{
		UserID:              uid,
		DislikedIngredients: disliked,
		MaxPrepMinutes:      *req.MaxPrepMinutes,
		PreferredCuisines:   cuisines,
	}

	updated, err := h.repo.Update(c.Request().Context(), uid, prefs)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to save preferences"})
	}
	return c.JSON(http.StatusOK, toPreferencesDTO(updated))
}

func toPreferencesDTO(p model.Preferences) model.PreferencesDTO {
	disliked := p.DislikedIngredients
	if disliked == nil {
		disliked = []string{}
	}
	cuisines := p.PreferredCuisines
	if cuisines == nil {
		cuisines = []string{}
	}
	return model.PreferencesDTO{
		UserID:              p.UserID,
		DislikedIngredients: disliked,
		MaxPrepMinutes:      p.MaxPrepMinutes,
		PreferredCuisines:   cuisines,
		UpdatedAt:           p.UpdatedAt,
	}
}