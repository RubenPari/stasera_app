package model

import (
	"time"

	"github.com/google/uuid"
)

// MealPlan represents a weekly dinner plan for a user.
type MealPlan struct {
	ID        uuid.UUID     `db:"id"         json:"id"`
	UserID    uuid.UUID     `db:"user_id"    json:"user_id"`
	WeekStart time.Time     `db:"week_start" json:"week_start"`
	Status    string        `db:"status"     json:"status"`
	Days      []MealPlanDay `json:"days"`
	CreatedAt time.Time     `db:"created_at" json:"created_at"`
}

// MealPlanDay links a day of the week to a recipe inside a plan.
type MealPlanDay struct {
	ID        uuid.UUID `db:"id"          json:"id"`
	PlanID    uuid.UUID `db:"plan_id"     json:"plan_id"`
	DayOfWeek int       `db:"day_of_week" json:"day_of_week"`
	RecipeID  uuid.UUID `db:"recipe_id"   json:"recipe_id"`
	Recipe    *Recipe   `db:"-" json:"recipe,omitempty"`
}
