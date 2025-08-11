package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/a-h/templ"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/gin-gonic/gin"
	"github.com/thomas-marquis/goLLMan/agent"
	"github.com/thomas-marquis/goLLMan/agent/session"
	"github.com/thomas-marquis/goLLMan/controller/server/components"
	"github.com/thomas-marquis/goLLMan/internal/domain"
	"github.com/thomas-marquis/goLLMan/pkg"
	"github.com/yuin/goldmark"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func (s *Server) FlowsHandlers(r *gin.Engine, g *genkit.Genkit) {
	for _, flow := range genkit.ListFlows(g) {
		r.POST("/"+flow.Name(), func(c *gin.Context) {
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
			c.HTML(http.StatusInternalServerError, "", components.Toast(
				components.ToastLevelError, "Books loading failed", err.Error()))
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
			c.HTML(http.StatusInternalServerError, "", components.Toast(
				components.ToastLevelError, "Session loading failed", err.Error()))
			return
		}

		var formData messageSubmitFormData
		if err := c.Bind(&formData); err != nil {
			c.HTML(http.StatusBadRequest, "", components.Toast(
				components.ToastLevelError, "Invalid format", err.Error()))
			return
		}

		// TODO: inject the agent instead the flow
		// TODO: pass the session ID to the agent class
		// TODO: dynamically set the book ID based on the current book context
		in := agent.ChatbotInput{
			Question: formData.Question,
			Session:  sess.ID(),
		}
		if _, err = s.flow.Run(c.Request.Context(), in); err != nil {
			pkg.Logger.Printf("Failed to generate response from flow: %s\n", err)
			c.HTML(http.StatusInternalServerError, "", components.Toast(
				components.ToastLevelError, "Response generation failed", err.Error()))
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
			c.HTML(http.StatusInternalServerError, "", components.Toast(
				components.ToastLevelError, "Session loading failed", err.Error()))
			return
		}
		stream.AttachSession(sess)

		v, ok := c.Get(clientChanKey)
		if !ok {
			pkg.Logger.Println("client chanel not found")
			c.HTML(http.StatusInternalServerError, "", components.Toast(
				components.ToastLevelError, "Ooops...", "Something went wrong. Please try again later."))
			return
		}
		clientMessageChan, ok := v.(messagesChan)
		if !ok {
			c.HTML(http.StatusInternalServerError, "", components.Toast(
				components.ToastLevelError, "Ooops...", "internal processing error"))
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
					c.HTML(http.StatusInternalServerError, "", components.Toast(
						components.ToastLevelError, "Response processing failed", err.Error()))
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

func (s *Server) NotificationHandlers(r *gin.Engine) {
	r.GET("notifications/end", func(c *gin.Context) {
		c.HTML(http.StatusOK, "", "")
	})
}

func (s *Server) UploadBookHandler(r *gin.Engine) {
	r.GET("books/upload/open", func(c *gin.Context) {
		c.HTML(http.StatusOK, "", components.UploadEpubModal())
	})

	r.GET("books/upload/cancel", func(c *gin.Context) {
		c.HTML(http.StatusOK, "", components.UploadEpubModalOff())
	})

	r.POST("/books/upload", func(c *gin.Context) {
		defer func() {
			c.HTML(http.StatusOK, "", components.UploadEpubModalOff())
		}()

		// Get the file from the request
		file, err := c.FormFile("epub-file")
		if err != nil {
			pkg.Logger.Printf("Error getting file sent by user: %v\n", err)
			c.HTML(http.StatusBadRequest, "", components.Toast(
				components.ToastLevelError, "Upload failed", "No file uploaded or invalid file"))
			return
		}

		// Verify the file is an EPUB
		if filepath.Ext(file.Filename) != ".epub" {
			pkg.Logger.Printf("Invalid file type: %s\n", file.Filename)
			c.HTML(http.StatusBadRequest, "", components.Toast(
				components.ToastLevelError, "Bad format", "only EPUB files are allowed"))
			return
		}

		// Create temporary directory if it doesn't exist
		tempDir := "./tmp/uploads"
		if err := os.MkdirAll(tempDir, 0755); err != nil {
			pkg.Logger.Printf("Error creating directory: %v", err)
			c.HTML(http.StatusInternalServerError, "", components.Toast(
				components.ToastLevelError, "Upload failed", "Failed to prepare for upload"))
			return
		}

		// Generate a unique filename to avoid collisions
		timestamp := time.Now().Unix()
		tempFilePath := fmt.Sprintf("%s/%d_%s", tempDir, timestamp, file.Filename)

		// Save the file to disk
		if err := c.SaveUploadedFile(file, tempFilePath); err != nil {
			pkg.Logger.Printf("Error saving file: %v", err)
			c.HTML(http.StatusInternalServerError, "", components.Toast(
				components.ToastLevelError, "Upload failed", "Failed to save file"))
			return
		}

		// Process the EPUB file
		book, err := s.bookRepository.ReadFromFile(c.Request.Context(), tempFilePath)
		if err != nil {
			pkg.Logger.Printf("Error processing EPUB: %v", err)
			os.Remove(tempFilePath)
			c.HTML(http.StatusInternalServerError, "", components.Toast(
				components.ToastLevelError, "Upload failed",
				"Failed to process EPUB file. Is your file corrupted?"))
			return
		}

		// Add the book to the repository
		addedBook, err := s.bookRepository.Add(
			c.Request.Context(),
			book.Title,
			book.Author,
			book.Metadata,
		)
		if err != nil {
			os.Remove(tempFilePath)
			if errors.Is(domain.ErrBookAlreadyExists, err) {
				c.HTML(http.StatusOK, "", components.Toast(
					components.ToastLevelInfo, "Book already exists",
					"A book with the same title and author already exists in the library"))
				return
			}
			pkg.Logger.Printf("Error adding book to repository: %v", err)
			c.HTML(http.StatusInternalServerError, "", components.Toast(
				components.ToastLevelError, "Upload failed", "Failed to add book to library"))
			return
		}

		os.Remove(tempFilePath)
		c.HTML(http.StatusOK, "", components.Toast(components.ToastLevelSuccess, "Upload",
			fmt.Sprintf("Book %s uploaded successfully. Indexing is starting...", addedBook.Title)))
		c.HTML(http.StatusOK, "", components.BookCard(addedBook))
	})
}

func sendToStream(c *gin.Context, comp templ.Component) {
	buff := new(bytes.Buffer)
	if err := comp.Render(c.Request.Context(), buff); err != nil {
		pkg.Logger.Printf("Error rendering component: %s\n", err)
		c.HTML(http.StatusInternalServerError, "", components.Toast(
			components.ToastLevelError, "Internal error", err.Error()))
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
