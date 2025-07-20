package server

import (
	"context"
	"errors"
	"github.com/thomas-marquis/goLLMan/agent/session"
)

const sessionID = "masession"

func getSession(store session.Store, ctx context.Context) (*session.Session, error) {
	sess, err := store.GetByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			sess, err = store.NewSession(ctx, session.WithLimit(10), session.WithID(sessionID))
		} else {
			return nil, err
		}
	}
	return sess, nil
}
