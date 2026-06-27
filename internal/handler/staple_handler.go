package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/stasera/stasera-api/internal/model"
)

// StapleHandler exposes endpoints to manage the user's staples.
type StapleHandler struct {
	repo StapleStore
}

// NewStapleHandler returns a new StapleHandler.
func NewStapleHandler(repo StapleStore) *StapleHandler {
	return &StapleHandler{repo: repo}
}

type createStapleRequest struct {
	Name string `json:"name" validate:"required,max=200"`
}

type updateStapleRequest struct {
	IsActive *bool `json:"is_active" validate:"required"`
}

// List returns all staples for the user, active and inactive.
func (h *StapleHandler) List(c echo.Context) error {
	uid := mustUserID(c)

	staples, err := h.repo.GetAllByUserID(c.Request().Context(), uid)
	if err != nil {
		return respondError(c, err)
	}

	dtos := make([]model.StapleDTO, len(staples))
	for i, s := range staples {
		dtos[i] = toStapleDTO(s)
	}
	return c.JSON(http.StatusOK, dtos)
}

// Create adds a new staple or reactivates an existing one with the same name.
func (h *StapleHandler) Create(c echo.Context) error {
	uid := mustUserID(c)

	var req createStapleRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	staple, err := h.repo.Create(c.Request().Context(), uid, req.Name)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(http.StatusCreated, toStapleDTO(staple))
}

// Update toggles the is_active flag of a staple owned by the user.
func (h *StapleHandler) Update(c echo.Context) error {
	uid := mustUserID(c)

	id, err := parsePathUUID(c, "id")
	if err != nil {
		return err
	}

	var req updateStapleRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	staple, err := h.repo.UpdateActive(c.Request().Context(), uid, id, *req.IsActive)
	if err != nil {
		return respondError(c, err)
	}
	if staple == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "staple not found"})
	}
	return c.JSON(http.StatusOK, toStapleDTO(*staple))
}

// Delete removes a staple owned by the user.
func (h *StapleHandler) Delete(c echo.Context) error {
	uid := mustUserID(c)

	id, err := parsePathUUID(c, "id")
	if err != nil {
		return err
	}

	deleted, err := h.repo.DeleteByID(c.Request().Context(), uid, id)
	if err != nil {
		return respondError(c, err)
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