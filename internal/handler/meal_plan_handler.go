package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/stasera/stasera-api/internal/model"
	"github.com/stasera/stasera-api/internal/service"
)

// MealPlanHandler handles meal plan endpoints.
type MealPlanHandler struct {
	service *service.MealPlanService
}

// NewMealPlanHandler returns a new MealPlanHandler.
func NewMealPlanHandler(svc *service.MealPlanService) *MealPlanHandler {
	return &MealPlanHandler{service: svc}
}

type swapDayRequest struct {
	RecipeID   *string `json:"recipe_id"`
	Regenerate *bool   `json:"regenerate"`
}

// Current returns the active meal plan for the current week.
func (h *MealPlanHandler) Current(c echo.Context) error {
	uid := mustUserID(c)

	plan, err := h.service.GetCurrent(c.Request().Context(), uid)
	if err != nil {
		return respondError(c, err)
	}
	if plan == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "no active meal plan"})
	}

	return c.JSON(http.StatusOK, toMealPlanDTO(*plan))
}

// Generate asks the AI to create a new weekly meal plan.
func (h *MealPlanHandler) Generate(c echo.Context) error {
	uid := mustUserID(c)

	plan, days, err := h.service.Generate(c.Request().Context(), uid)
	if err != nil {
		return respondError(c, err)
	}

	dto := toMealPlanDTO(plan)
	dto.Days = make([]model.MealPlanDayDTO, len(days))
	for i, d := range days {
		dto.Days[i] = toMealPlanDayDTO(d)
	}
	return c.JSON(http.StatusCreated, dto)
}

// SwapDay replaces the recipe for a specific day.
func (h *MealPlanHandler) SwapDay(c echo.Context) error {
	uid := mustUserID(c)

	planID, err := parsePathUUID(c, "planId")
	if err != nil {
		return err
	}

	dayOfWeek, err := parsePathInt(c, "dayOfWeek", 1, 5)
	if err != nil {
		return err
	}

	var req swapDayRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	if req.Regenerate != nil && *req.Regenerate {
		day, err := h.service.RegenerateDay(c.Request().Context(), uid, planID, dayOfWeek)
		if err != nil {
			return respondError(c, err)
		}
		return c.JSON(http.StatusOK, toMealPlanDayDTO(day))
	}

	if req.RecipeID == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "recipe_id or regenerate is required"})
	}
	recipeID, err := uuid.Parse(*req.RecipeID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid recipe id"})
	}

	day, err := h.service.SwapDay(c.Request().Context(), uid, planID, dayOfWeek, recipeID)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(http.StatusOK, toMealPlanDayDTO(day))
}

// Today returns the recipe assigned to today in the active plan.
func (h *MealPlanHandler) Today(c echo.Context) error {
	uid := mustUserID(c)

	recipe, err := h.service.GetToday(c.Request().Context(), uid)
	if err != nil {
		return respondError(c, err)
	}
	if recipe == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "no recipe planned for today"})
	}

	return c.JSON(http.StatusOK, toRecipeDTO(*recipe))
}

func toMealPlanDTO(p model.MealPlan) model.MealPlanDTO {
	dto := model.MealPlanDTO{
		ID:        p.ID,
		UserID:    p.UserID,
		WeekStart: p.WeekStart,
		Status:    p.Status,
		CreatedAt: p.CreatedAt,
		Days:      make([]model.MealPlanDayDTO, 0, len(p.Days)),
	}
	for _, d := range p.Days {
		dto.Days = append(dto.Days, toMealPlanDayDTO(d))
	}
	return dto
}

func toMealPlanDayDTO(d model.MealPlanDay) model.MealPlanDayDTO {
	dto := model.MealPlanDayDTO{
		ID:        d.ID,
		PlanID:    d.PlanID,
		DayOfWeek: d.DayOfWeek,
		RecipeID:  d.RecipeID,
	}
	if d.Recipe != nil {
		r := toRecipeDTO(*d.Recipe)
		dto.Recipe = &r
	}
	return dto
}