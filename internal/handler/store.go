package handler

import (
	"context"

	"github.com/google/uuid"
	"github.com/stasera/stasera-api/internal/model"
)

// RecipeStore is the recipe persistence surface consumed by RecipeHandler.
type RecipeStore interface {
	GetByIDForUser(ctx context.Context, id, userID uuid.UUID) (*model.Recipe, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, isRescue *bool) ([]model.Recipe, error)
	MarkCookedForUser(ctx context.Context, id, userID uuid.UUID) (*model.Recipe, error)
	DeleteForUser(ctx context.Context, id, userID uuid.UUID) (bool, error)
}

// StapleStore is the staple persistence surface consumed by StapleHandler.
type StapleStore interface {
	GetAllByUserID(ctx context.Context, userID uuid.UUID) ([]model.Staple, error)
	Create(ctx context.Context, userID uuid.UUID, name string) (model.Staple, error)
	UpdateActive(ctx context.Context, userID, id uuid.UUID, isActive bool) (*model.Staple, error)
	DeleteByID(ctx context.Context, userID, id uuid.UUID) (bool, error)
}

// PreferencesStore is the preferences persistence surface consumed by PreferencesHandler.
type PreferencesStore interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (*model.Preferences, error)
	Update(ctx context.Context, userID uuid.UUID, p model.Preferences) (model.Preferences, error)
}

// UserStore is the user persistence surface consumed by AuthHandler.
type UserStore interface {
	Create(ctx context.Context, email, passwordHash, displayName string) (model.User, error)
	CreateWithDefaultStaples(ctx context.Context, email, passwordHash, displayName string) (model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	UpdateProfile(ctx context.Context, id uuid.UUID, displayName string) (model.User, error)
	UpdatePasswordHash(ctx context.Context, id uuid.UUID, hash string) error
}