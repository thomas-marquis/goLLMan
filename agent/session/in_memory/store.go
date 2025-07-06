package in_memory

import (
	"context"
	"github.com/thomas-marquis/goLLMan/agent/session"
	"sync"
)

type InMemorySessionStore struct {
	sync.Mutex
	sessions map[string]*session.Session
}

var _ session.Store = (*InMemorySessionStore)(nil)

func NewSessionStore() *InMemorySessionStore {
	return &InMemorySessionStore{
		sessions: make(map[string]*session.Session),
	}
}

func (i *InMemorySessionStore) Save(ctx context.Context, sess *session.Session) error {
	i.Lock()
	defer i.Unlock()
	i.sessions[sess.ID()] = sess
	return nil
}

func (i *InMemorySessionStore) GetByID(ctx context.Context, id string) (*session.Session, error) {
	i.Lock()
	defer i.Unlock()
	sess, ok := i.sessions[id]
	if !ok {
		return nil, session.ErrSessionNotFound
	}
	return sess, nil
}
