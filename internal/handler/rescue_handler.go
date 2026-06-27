package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/stasera/stasera-api/internal/model"
	"github.com/stasera/stasera-api/internal/service"
)

// RescueHandler exposes the AI rescue-meals endpoint as a thin layer over
// RescueService. It returns a bare JSON array of recipes (no envelope).
type RescueHandler struct {
	svc *service.RescueService
}

// NewRescueHandler returns a new RescueHandler.
func NewRescueHandler(svc *service.RescueService) *RescueHandler {
	return &RescueHandler{svc: svc}
}

// Rescue generates 3 emergency meals using the user's active staples.
// Response: bare array []model.RecipeDTO.
func (h *RescueHandler) Rescue(c echo.Context) error {
	uid := mustUserID(c)

	recipes, err := h.svc.Generate(c.Request().Context(), uid)
	if err != nil {
		return respondError(c, err)
	}

	dtos := make([]model.RecipeDTO, len(recipes))
	for i, r := range recipes {
		dtos[i] = toRecipeDTO(r)
	}
	return c.JSON(http.StatusOK, dtos)
}