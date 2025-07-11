package server

import (
	"github.com/firebase/genkit/go/ai"
	"github.com/thomas-marquis/goLLMan/agent/session"
	"github.com/thomas-marquis/goLLMan/pkg"
)

// TODO: this type cloud be owned by the agentic layer
type messagesChan chan *ai.Message

type eventStream struct {
	// Events are pushed to this channel by the main events-gathering routine
	MessageBySessionID map[string]messagesChan

	// New client connections
	NewClients chan messagesChan

	// Closed client connections
	ClosedClients chan messagesChan

	// Total client connections
	TotalClients map[messagesChan]struct{}
}

func (s *eventStream) AttachSession(sess *session.Session) {
	if _, exists := s.MessageBySessionID[sess.ID()]; exists {
		pkg.Logger.Printf("Session %s already attached, skipping", sess.ID())
		return
	}
	s.MessageBySessionID[sess.ID()] = sess.ListenMessages()

	go func() {
		for {
			select {
			case msg := <-s.MessageBySessionID[sess.ID()]:
				// Broadcast the message to all clients
				for clientMessageChan := range s.TotalClients {
					select {
					case clientMessageChan <- msg:
					default:
						pkg.Logger.Printf("Client message channel full, dropping message")
					}
				}
			}
		}
	}()
}

// It Listens all incoming requests from clients.
// Handles addition and removal of clients and broadcast messages to clients.
func (s *eventStream) listen() {
	for {
		select {
		// Add new available client
		case client := <-s.NewClients:
			s.TotalClients[client] = struct{}{}
			pkg.Logger.Printf("Client added. %d registered clients", len(s.TotalClients))

		// Remove closed client
		case client := <-s.ClosedClients:
			delete(s.TotalClients, client)
			close(client)
			pkg.Logger.Printf("Removed client. %d registered clients", len(s.TotalClients))
		}
	}
}
