package server

import (
	"bytes"
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

func (s *Server) GetPageHandler(r *gin.Engine) {
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "", components.Page())
	})
}

func (s *Server) PostMessageHandler(r *gin.Engine, store session.Store) {
	r.POST("/messages", func(c *gin.Context) {
		pkg.Logger.Println("MessageBySessionID received")

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
		_, err = s.flow.Run(c.Request.Context(), agent.ChatbotInput{Question: formData.Question, Session: sess.ID()})
		if err != nil {
			pkg.Logger.Printf("EFailed to generate response from flow: %s\n", err)
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
		stream.AttachSession(sess)

		v, ok := c.Get(clientChanKey)
		if !ok {
			c.HTML(http.StatusInternalServerError, "",
				components.ErrorBanner("client not found"))
			return
		}
		clientChan, ok := v.(messagesChan)
		if !ok {
			c.HTML(http.StatusInternalServerError, "",
				components.ErrorBanner("internal processing error"))
			return
		}

		go func() {
			pkg.Logger.Println("Catch up previous messages:")
			for _, m := range sess.GetMessages() {
				pkg.Logger.Println(pkg.ContentToText(m.Content))
				clientChan <- m
			}
		}()

		c.Stream(func(w io.Writer) bool {
			if msg, ok := <-clientChan; ok {
				_, err := convertToHTML(pkg.ContentToText(msg.Content))
				if err != nil {
					pkg.Logger.Printf("Error converting content to HTML: %s\n", err)
					c.HTML(http.StatusInternalServerError, "", components.ErrorBanner(err.Error()))
					return false
				}
				buff := new(bytes.Buffer)
				components.Message(
					string(msg.Role),
					pkg.ContentToText(msg.Content),
				).Render(c.Request.Context(), buff)
				c.SSEvent("message", buff.String())
				return true
			}
			return false
		})
	})
}

func convertToHTML(content string) (string, error) {
	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(content), &buf); err != nil {
		return "", err
	}

	return buf.String(), nil
}
