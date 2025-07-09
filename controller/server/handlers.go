package server

import (
	"bytes"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/gin-gonic/gin"
	"github.com/thomas-marquis/goLLMan/agent/session"
	"github.com/thomas-marquis/goLLMan/controller/server/components"
	"github.com/thomas-marquis/goLLMan/pkg"
	"io"
	"net/http"
	"time"
)

func (s *Server) FlowsHandlers(r *gin.Engine, g *genkit.Genkit) {
	for _, flow := range genkit.ListFlows(g) {
		s.router.POST("/"+flow.Name(), func(c *gin.Context) {
			genkit.Handler(flow)(c.Writer, c.Request)
		})
	}
}

func (s *Server) GetPageHandler(r *gin.Engine) {
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "", components.Page())
	})
}

func (s *Server) PostMessageHandler(r *gin.Engine, store session.Store) {
	r.POST("/messages", func(c *gin.Context) {
		pkg.Logger.Println("Message received")

		sess, err := getSession(store, c.Request.Context())
		if err != nil {
			pkg.Logger.Println(err)
			c.HTML(http.StatusInternalServerError, "", components.ErrorBanner(err.Error()))
			return
		}

		var formData messageSubmitFormData
		if err := c.Bind(&formData); err != nil {
			c.HTML(http.StatusInternalServerError, "", components.ErrorBanner(err.Error()))
			return
		}
		msg := ai.NewMessage(ai.RoleUser, nil, pkg.ContentFromText(formData.Question)...)
		if err := sess.AddMessage(msg); err != nil {
			pkg.Logger.Printf("Error adding message to session: %s\n", err)
			c.HTML(http.StatusInternalServerError, "", components.ErrorBanner(err.Error()))
			return
		}

		time.Sleep(2 * time.Second) // Simulate processing delay
		response := "Je suis un chatbot, je ne peux pas répondre à cette question."
		botResponse := ai.NewMessage(ai.RoleModel, nil, pkg.ContentFromText(response)...)
		if err := sess.AddMessage(botResponse); err != nil {
			pkg.Logger.Printf("Error adding message to session: %s\n", err)
			c.HTML(http.StatusInternalServerError, "", components.ErrorBanner(err.Error()))
			return
		}

		c.HTML(http.StatusOK, "", components.Message("user", formData.Question))
	})
}

func (s *Server) SSEMessagesHandler(r *gin.Engine, store session.Store, stream *eventStream) {
	r.GET("/stream", headersSSEMiddleware(), stream.sseConnectMiddleware(), func(c *gin.Context) {
		pkg.Logger.Println("Streaming messages")
		sess, err := getSession(store, c.Request.Context())
		if err != nil {
			pkg.Logger.Println(err)
			c.HTML(http.StatusInternalServerError, "", components.ErrorBanner(err.Error()))
			return
		}

		v, ok := c.Get(clientChanKey)
		if !ok {
			return
		}
		clientChan, ok := v.(messagesChan)
		if !ok {
			return
		}
		sess.Sub(clientChan)

		c.Stream(func(w io.Writer) bool {
			// Stream message to client from message channel
			if msg, ok := <-clientChan; ok {
				buff := new(bytes.Buffer)
				components.Message(string(msg.Role),
					pkg.ContentToText(msg.Content)).Render(c.Request.Context(), buff)
				c.SSEvent("message", buff.String())
				return true
			}
			return false
		})
	})
}
