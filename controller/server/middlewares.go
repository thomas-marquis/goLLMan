package server

import "github.com/gin-gonic/gin"

const (
	clientChanKey = "messagesChan"
)

func headersSSEMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Writer.Header().Set("Transfer-Encoding", "chunked")
		c.Next()
	}
}

func (s *eventStream) sseConnectMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Initialize client channel
		clientChan := make(messagesChan, 1)

		// Send new connection to event server
		s.NewClients <- clientChan

		go func() {
			<-c.Writer.CloseNotify()

			// Drain client channel so that it does not block. Server may keep sending messages to this channel
			for range clientChan {
			}
			// Send closed connection to event server
			s.ClosedClients <- clientChan
		}()

		c.Set(clientChanKey, clientChan)

		c.Next()
	}
}
