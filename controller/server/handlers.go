package server

import (
	"bytes"
	"context"
	"github.com/a-h/templ"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/gin-gonic/gin"
	"github.com/thomas-marquis/goLLMan/agent"
	"github.com/thomas-marquis/goLLMan/agent/session"
	"github.com/thomas-marquis/goLLMan/controller/server/components"
	"github.com/thomas-marquis/goLLMan/pkg"
	"github.com/yuin/goldmark"
	"io"
	"net/http"
)

func (s *Server) FlowsHandlers(r *gin.Engine, g *genkit.Genkit) {
	for _, flow := range genkit.ListFlows(g) {
		s.router.POST("/"+flow.Name(), func(c *gin.Context) {
			genkit.Handler(flow)(c.Writer, c.Request)
		})
	}
}

func (s *Server) ToggleBookSelectionHandler(r *gin.Engine) {
	r.POST("/books/:id/toggle", func(c *gin.Context) {
		// TODO: save change here
		pkg.Logger.Printf("Toggle book selection: %s\n", c.Param("id"))
		c.Status(http.StatusOK)
	})
}

func (s *Server) GetPageHandler(r *gin.Engine) {
	r.GET("/", func(c *gin.Context) {
		books, err := s.bookRepository.List(context.Background())
		if err != nil {
			pkg.Logger.Printf("Failed to list books: %s\n", err)
			c.HTML(http.StatusInternalServerError, "", components.ErrorBanner(err.Error()))
			return
		}
		c.HTML(http.StatusOK, "", components.Page(books))
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

		// TODO: inject the agent instead the flow
		// TODO: pass the session ID to the agent class
		// TODO: dynamically set the book ID based on the current book context
		in := agent.ChatbotInput{
			Question: formData.Question,
			Session:  sess.ID(),
			BookID:   "1",
		}
		if _, err = s.flow.Run(c.Request.Context(), in); err != nil {
			pkg.Logger.Printf("Failed to generate response from flow: %s\n", err)
			c.HTML(http.StatusInternalServerError, "", components.ErrorBanner(err.Error()))
			return
		}

		c.HTML(http.StatusOK, "",
			components.Message("user", formData.Question),
		)
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
		stream.AttachSession(sess)

		v, ok := c.Get(clientChanKey)
		if !ok {
			c.HTML(http.StatusInternalServerError, "",
				components.ErrorBanner("client not found"))
			return
		}
		clientMessageChan, ok := v.(messagesChan)
		if !ok {
			c.HTML(http.StatusInternalServerError, "",
				components.ErrorBanner("internal processing error"))
			return
		}

		go func() {
			for _, m := range sess.GetMessages() {
				pkg.Logger.Println(pkg.ContentToText(m.Content))
				clientMessageChan <- m
			}
		}()

		c.Stream(func(w io.Writer) bool {
			if msg, ok := <-clientMessageChan; ok {
				if msg.Role == ai.RoleUser {
					sendToStream(c, components.Thinking())
				} else {
					sendToStream(c, components.NotThinking())
				}

				if _, err := convertToHTML(pkg.ContentToText(msg.Content)); err != nil {
					pkg.Logger.Printf("Error converting content to HTML: %s\n", err)
					c.HTML(http.StatusInternalServerError, "", components.ErrorBanner(err.Error()))
					return false
				}

				sendToStream(c,
					components.Message(
						string(msg.Role),
						pkg.ContentToText(msg.Content),
					))
				return true
			}
			return false
		})
	})
}

func sendToStream(c *gin.Context, comp templ.Component) {
	buff := new(bytes.Buffer)
	if err := comp.Render(c.Request.Context(), buff); err != nil {
		pkg.Logger.Printf("Error rendering component: %s\n", err)
		c.HTML(http.StatusInternalServerError, "", components.ErrorBanner(err.Error()))
		return
	}
	c.SSEvent("message", buff.String())
}

func convertToHTML(content string) (string, error) {
	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(content), &buf); err != nil {
		return "", err
	}

	return buf.String(), nil
}
