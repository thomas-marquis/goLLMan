package session

import (
	"context"
	"errors"
)

var (
	ErrSessionNotFound = errors.New("session not found")
)

type Store interface {
	// NewSession creates a new session with a unique ID.
	NewSession(ctx context.Context, opts ...Option) (*Session, error)

	// GetByID retrieves a session by its ID.
	// If the session does not exist, it returns ErrSessionNotFound.
	GetByID(ctx context.Context, id string) (*Session, error)
}
