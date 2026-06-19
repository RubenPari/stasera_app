package model

import (
	"github.com/google/uuid"
)

// Staple is an ingredient the user always keeps at home.
type Staple struct {
	ID       uuid.UUID `db:"id"         json:"id"`
	UserID   uuid.UUID `db:"user_id"    json:"user_id"`
	Name     string    `db:"name"       json:"name"`
	IsActive bool      `db:"is_active"  json:"is_active"`
}
