package model

import (
	"time"

	"github.com/google/uuid"
)

// ShoppingItem is a single product entry in a shopping list.
type ShoppingItem struct {
	ID        uuid.UUID `db:"id"         json:"id"`
	ListID    uuid.UUID `db:"list_id"    json:"list_id"`
	Name      string    `db:"name"       json:"name"`
	Quantity  string    `db:"quantity"   json:"quantity"`
	Aisle     string    `db:"aisle"      json:"aisle"`
	IsChecked bool      `db:"is_checked" json:"is_checked"`
	SortOrder int       `db:"sort_order" json:"sort_order"`
}

// ShoppingList aggregates items generated from a meal plan.
type ShoppingList struct {
	ID          uuid.UUID      `db:"id"            json:"id"`
	UserID      uuid.UUID      `db:"user_id"       json:"user_id"`
	PlanID      *uuid.UUID     `db:"plan_id"       json:"plan_id,omitempty"`
	Items       []ShoppingItem `json:"items"`
	CreatedAt   time.Time      `db:"created_at"    json:"created_at"`
	CompletedAt *time.Time     `db:"completed_at"  json:"completed_at,omitempty"`
}
