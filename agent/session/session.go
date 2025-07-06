package session

import (
	"fmt"
	"github.com/firebase/genkit/go/ai"
	"github.com/google/uuid"
)

type Option func(s *Session)

func WithLimit(limit int) Option {
	if limit < 0 {
		limit = 0
	}
	if limit <= 2 {
		limit = 2
	}
	return func(s *Session) {
		if limit != 0 {
			s.limited = true
		} else {
			s.limited = false
		}
		s.limit = limit
	}
}

func WithID(id string) Option {
	return func(s *Session) {
		s.id = id
	}
}

func GenerateID() string {
	return uuid.New().String()
}

type Session struct {
	id               string
	messages         []*ai.Message
	limited          bool
	limit            int
	hasSystemMessage bool
}

func New(opts ...Option) *Session {
	s := &Session{
		messages:         make([]*ai.Message, 0),
		hasSystemMessage: false,
	}

	for _, opt := range opts {
		opt(s)
	}

	if s.id == "" {
		s.id = GenerateID()
	}

	return s
}

func (s *Session) ID() string {
	return s.id
}

func (s *Session) AddMessage(msg *ai.Message) error {
	if msg.Role == ai.RoleSystem {
		if s.hasSystemMessage {
			return fmt.Errorf("cannot add message to system message")
		}
		if len(s.messages) > 0 {
			return fmt.Errorf("system message must be the first message in the session")
		}
		s.hasSystemMessage = true
		s.limit += 1
	}

	s.messages = append(s.messages, msg)

	if s.limited && len(s.messages) > s.limit {
		if s.hasSystemMessage {
			s.messages = append([]*ai.Message{s.messages[0]}, s.messages[2:]...)
		} else {
			s.messages = s.messages[1:]
		}
	}

	return nil
}

func (s *Session) GetMessages() ([]*ai.Message, error) {
	return s.messages, nil
}
