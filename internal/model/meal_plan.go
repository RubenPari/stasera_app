package model

import (
	"time"

	"github.com/google/uuid"
)

// MealPlan represents a weekly dinner plan for a user.
type MealPlan struct {
	ID        uuid.UUID     `json:"id"`
	UserID    uuid.UUID     `json:"user_id"`
	WeekStart time.Time     `json:"week_start"`
	Status    string        `json:"status"`
	Days      []MealPlanDay `json:"days"`
	CreatedAt time.Time     `json:"created_at"`
}

// MealPlanDay links a day of the week to a recipe inside a plan.
type MealPlanDay struct {
	ID        uuid.UUID `json:"id"`
	PlanID    uuid.UUID `json:"plan_id"`
	DayOfWeek int       `json:"day_of_week"`
	RecipeID  uuid.UUID `json:"recipe_id"`
	Recipe    *Recipe   `json:"recipe,omitempty"`
}