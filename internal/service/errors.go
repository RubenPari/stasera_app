package service

import "errors"

// Sentinel errors for the service layer. Business failures are wrapped around
// these sentinels (e.g. fmt.Errorf("load preferences: %w", ErrNotFound)) so the
// handler layer can map them to HTTP status codes via errors.Is.
var (
	ErrNotFound      = errors.New("not found")
	ErrConflict      = errors.New("conflict")
	ErrValidation    = errors.New("validation")
	ErrAIUnavailable = errors.New("ai unavailable")
)