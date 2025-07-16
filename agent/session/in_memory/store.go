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

func (s *InMemorySessionStore) NewSession(ctx context.Context, opts ...session.Option) (*session.Session, error) {
	s.Lock()
	defer s.Unlock()
	opts = append([]session.Option{session.WithID(session.GenerateID())}, opts...)
	sess := session.New(opts...)
	s.sessions[sess.ID()] = sess
	return sess, nil
}

func (s *InMemorySessionStore) Save(ctx context.Context, sess *session.Session) error {
	s.Lock()
	defer s.Unlock()
	s.sessions[sess.ID()] = sess
	return nil
}

func (s *InMemorySessionStore) GetByID(ctx context.Context, id string) (*session.Session, error) {
	s.Lock()
	defer s.Unlock()
	sess, ok := s.sessions[id]
	if !ok {
		return nil, session.ErrSessionNotFound
	}
	return sess, nil
}
