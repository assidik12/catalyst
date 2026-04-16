package domain

import "errors"

// Sentinel errors for the domain layer.
// Service methods wrap these with fmt.Errorf("%w: ...", domain.ErrXxx) so
// callers can use errors.Is() to distinguish error categories without
// depending on error string comparisons.
var (
	// ErrNotFound is returned when a requested resource does not exist.
	ErrNotFound = errors.New("resource not found")

	// ErrInvalidInput is returned when caller-supplied data fails validation.
	ErrInvalidInput = errors.New("invalid input")

	// ErrUnauthorized is returned when the caller lacks permission.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrConflict is returned when a resource already exists (e.g. duplicate email).
	ErrConflict = errors.New("resource already exists")
)
