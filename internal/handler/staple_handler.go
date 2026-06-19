package handler

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/stasera/stasera-api/internal/middleware"
	"github.com/stasera/stasera-api/internal/model"
	"github.com/stasera/stasera-api/internal/repository"
)

// StapleHandler exposes endpoints to manage the user's staples.
type StapleHandler struct {
	repo      *repository.StapleRepository
	validator *validator.Validate
}

// NewStapleHandler returns a new StapleHandler.
func NewStapleHandler(repo *repository.StapleRepository) *StapleHandler {
	return &StapleHandler{repo: repo, validator: validator.New()}
}

type createStapleRequest struct {
	Name string `json:"name" validate:"required,max=200"`
}

type updateStapleRequest struct {
	IsActive *bool `json:"is_active" validate:"required"`
}

// List returns all staples for the user, active and inactive.
func (h *StapleHandler) List(c echo.Context) error {
	uid, err := middleware.GetUserID(c)
	if err != nil {
		return middleware.Unauthorized(c, "missing or invalid token")
	}

	staples, err := h.repo.GetAllByUserID(c.Request().Context(), uid)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to load staples"})
	}

	dtos := make([]model.StapleDTO, len(staples))
	for i, s := range staples {
		dtos[i] = toStapleDTO(s)
	}
	return c.JSON(http.StatusOK, dtos)
}

// Create adds a new staple or reactivates an existing one with the same name.
func (h *StapleHandler) Create(c echo.Context) error {
	uid, err := middleware.GetUserID(c)
	if err != nil {
		return middleware.Unauthorized(c, "missing or invalid token")
	}

	var req createStapleRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	if err := h.validator.Struct(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	staple, err := h.repo.Create(c.Request().Context(), uid, req.Name)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create staple"})
	}
	return c.JSON(http.StatusCreated, toStapleDTO(staple))
}

// Update toggles the is_active flag of a staple owned by the user.
func (h *StapleHandler) Update(c echo.Context) error {
	uid, err := middleware.GetUserID(c)
	if err != nil {
		return middleware.Unauthorized(c, "missing or invalid token")
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid staple id"})
	}

	var req updateStapleRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	if err := h.validator.Struct(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	staple, err := h.repo.UpdateActive(c.Request().Context(), uid, id, *req.IsActive)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update staple"})
	}
	if staple == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "staple not found"})
	}
	return c.JSON(http.StatusOK, toStapleDTO(*staple))
}

// Delete removes a staple owned by the user.
func (h *StapleHandler) Delete(c echo.Context) error {
	uid, err := middleware.GetUserID(c)
	if err != nil {
		return middleware.Unauthorized(c, "missing or invalid token")
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid staple id"})
	}

	deleted, err := h.repo.DeleteByID(c.Request().Context(), uid, id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to delete staple"})
	}
	if !deleted {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "staple not found"})
	}
	return c.NoContent(http.StatusNoContent)
}

func toStapleDTO(s model.Staple) model.StapleDTO {
	return model.StapleDTO{
		ID:       s.ID,
		UserID:   s.UserID,
		Name:     s.Name,
		IsActive: s.IsActive,
	}
}