package handler

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/stasera/stasera-api/internal/middleware"
	"github.com/stasera/stasera-api/internal/model"
	"github.com/stasera/stasera-api/internal/service"
)

// ShoppingListHandler exposes endpoints to generate and manage shopping lists.
type ShoppingListHandler struct {
	service *service.ShoppingListService
}

// NewShoppingListHandler returns a new ShoppingListHandler.
func NewShoppingListHandler(service *service.ShoppingListService) *ShoppingListHandler {
	return &ShoppingListHandler{service: service}
}

type updateItemRequest struct {
	IsChecked *bool `json:"is_checked" validate:"required"`
}

// Current returns the user's open shopping list with its items.
func (h *ShoppingListHandler) Current(c echo.Context) error {
	uid, err := middleware.GetUserID(c)
	if err != nil {
		return middleware.Unauthorized(c, "missing or invalid token")
	}

	list, err := h.service.GetCurrent(c.Request().Context(), uid)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	if list == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "no open shopping list"})
	}
	return c.JSON(http.StatusOK, toShoppingListDTO(*list))
}

// Generate aggregates the current meal plan into a new shopping list.
func (h *ShoppingListHandler) Generate(c echo.Context) error {
	uid, err := middleware.GetUserID(c)
	if err != nil {
		return middleware.Unauthorized(c, "missing or invalid token")
	}

	list, err := h.service.Generate(c.Request().Context(), uid)
	if err != nil {
		if errors.Is(err, service.ErrNoActivePlan) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, toShoppingListDTO(*list))
}

// UpdateItem toggles an item's checked state.
func (h *ShoppingListHandler) UpdateItem(c echo.Context) error {
	uid, err := middleware.GetUserID(c)
	if err != nil {
		return middleware.Unauthorized(c, "missing or invalid token")
	}

	itemID, err := uuid.Parse(c.Param("itemId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid item id"})
	}

	var req updateItemRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	if req.IsChecked == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "is_checked is required"})
	}

	item, err := h.service.UpdateItem(c.Request().Context(), uid, itemID, *req.IsChecked)
	if err != nil {
		if errors.Is(err, service.ErrItemNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, toShoppingItemDTO(*item))
}

// Complete marks the user's open shopping list as completed.
func (h *ShoppingListHandler) Complete(c echo.Context) error {
	uid, err := middleware.GetUserID(c)
	if err != nil {
		return middleware.Unauthorized(c, "missing or invalid token")
	}

	list, err := h.service.Complete(c.Request().Context(), uid)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	if list == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "no open shopping list"})
	}
	return c.JSON(http.StatusOK, toShoppingListDTO(*list))
}

func toShoppingListDTO(l model.ShoppingList) model.ShoppingListDTO {
	dto := model.ShoppingListDTO{
		ID:          l.ID,
		UserID:      l.UserID,
		PlanID:      l.PlanID,
		CreatedAt:   l.CreatedAt,
		CompletedAt: l.CompletedAt,
		Items:       make([]model.ShoppingItemDTO, 0, len(l.Items)),
	}
	for _, it := range l.Items {
		dto.Items = append(dto.Items, toShoppingItemDTO(it))
	}
	return dto
}

func toShoppingItemDTO(i model.ShoppingItem) model.ShoppingItemDTO {
	return model.ShoppingItemDTO{
		ID:        i.ID,
		ListID:    i.ListID,
		Name:      i.Name,
		Quantity:  i.Quantity,
		Aisle:     i.Aisle,
		IsChecked: i.IsChecked,
		SortOrder: i.SortOrder,
	}
}