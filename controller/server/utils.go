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
			sess = session.New(session.WithID(sessionID), session.WithLimit(10))
			if err := store.Save(ctx, sess); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return sess, nil
}
