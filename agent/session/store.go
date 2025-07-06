package session

import (
	"context"
	"errors"
)

var (
	ErrSessionNotFound = errors.New("session not found")
)

type Store interface {
	Save(ctx context.Context, sess *Session) error

	// GetByID retrieves a session by its ID.
	// If the session does not exist, it returns ErrSessionNotFound.
	GetByID(ctx context.Context, id string) (*Session, error)
}
