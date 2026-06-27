package model

import (
	"time"

	"github.com/google/uuid"
)

// ShoppingItem is a single product entry in a shopping list.
type ShoppingItem struct {
	ID        uuid.UUID `json:"id"`
	ListID    uuid.UUID `json:"list_id"`
	Name      string    `json:"name"`
	Quantity  string    `json:"quantity"`
	Aisle     string    `json:"aisle"`
	IsChecked bool      `json:"is_checked"`
	SortOrder int       `json:"sort_order"`
}

// ShoppingList aggregates items generated from a meal plan.
type ShoppingList struct {
	ID          uuid.UUID      `json:"id"`
	UserID      uuid.UUID      `json:"user_id"`
	PlanID      *uuid.UUID     `json:"plan_id,omitempty"`
	Items       []ShoppingItem `json:"items"`
	CreatedAt   time.Time      `json:"created_at"`
	CompletedAt *time.Time     `json:"completed_at,omitempty"`
}