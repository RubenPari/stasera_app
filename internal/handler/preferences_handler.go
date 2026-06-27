package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/stasera/stasera-api/internal/model"
)

// PreferencesHandler exposes endpoints to read and update dietary preferences.
type PreferencesHandler struct {
	repo PreferencesStore
}

// NewPreferencesHandler returns a new PreferencesHandler.
func NewPreferencesHandler(repo PreferencesStore) *PreferencesHandler {
	return &PreferencesHandler{repo: repo}
}

type updatePreferencesRequest struct {
	DislikedIngredients []string `json:"disliked_ingredients"`
	MaxPrepMinutes      *int     `json:"max_prep_minutes" validate:"required,min=1,max=120"`
	PreferredCuisines   []string `json:"preferred_cuisines"`
}

// Get returns the user's preferences. Returns 404 when none are set yet.
func (h *PreferencesHandler) Get(c echo.Context) error {
	uid := mustUserID(c)

	prefs, err := h.repo.GetByUserID(c.Request().Context(), uid)
	if err != nil {
		return respondError(c, err)
	}
	if prefs == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "preferences not found"})
	}
	return c.JSON(http.StatusOK, toPreferencesDTO(*prefs))
}

// Update replaces the user's preferences with the provided values.
func (h *PreferencesHandler) Update(c echo.Context) error {
	uid := mustUserID(c)

	var req updatePreferencesRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
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
		return respondError(c, err)
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