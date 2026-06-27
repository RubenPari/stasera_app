package model

import "github.com/google/uuid"

// Staple is an ingredient the user always keeps at home.
type Staple struct {
	ID       uuid.UUID `json:"id"`
	UserID   uuid.UUID `json:"user_id"`
	Name     string    `json:"name"`
	IsActive bool      `json:"is_active"`
}