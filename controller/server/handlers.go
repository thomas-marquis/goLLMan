package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"time"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/gin-gonic/gin"
	"github.com/thomas-marquis/goLLMan/agent"
	"github.com/thomas-marquis/goLLMan/agent/session"
	"github.com/thomas-marquis/goLLMan/controller/server/components"
	"github.com/thomas-marquis/goLLMan/internal/domain"
	"github.com/thomas-marquis/goLLMan/pkg"
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
		bookId := c.Param("id")

		book, err := s.bookRepository.GetByID(c.Request.Context(), bookId)
		if err != nil {
			pkg.Logger.Printf("Error getting book: %s\n", err)
			showError(c, err, "Internal error", "Unable to find this book")
			return
		}

		book.Selected = !book.Selected
		action := "selected"
		if !book.Selected {
			action = "unselected"
		}

		if err := s.bookRepository.Update(c.Request.Context(), book); err != nil {
			pkg.Logger.Printf("Error updating book: %s\n", err)
			showError(c, err, "Internal error", "Unable to select/unselect this book")
			return
		}

		pkg.Logger.Printf("Toggle book selection: %s\n", c.Param("id"))
		showSuccess(c, "Saved", "'%s' %s", book.Title, action)
	})
}

func (s *Server) GetPageHandler(r *gin.Engine) {
	r.GET("/", func(c *gin.Context) {
		books, err := s.bookRepository.List(context.Background())
		if err != nil {
			pkg.Logger.Printf("Failed to list books: %s\n", err)
			showError(c, err, "Books loading failed", "")
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
			showError(c, err, "Session loading failed", "")
			return
		}

		var formData messageSubmitFormData
		if err := c.Bind(&formData); err != nil {
			showError(c, err, "Invalid format", "")
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
			showError(c, err, "Response generation failed", "")
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
			showError(c, err, "Session loading failed", "")
			return
		}
		stream.AttachSession(sess)

		v, ok := c.Get(clientChanKey)
		if !ok {
			pkg.Logger.Println("client chanel not found")
			showError(c, nil, "Ooops...", "Something went wrong. Please try again later.")
			return
		}
		clientMessageChan, ok := v.(messagesChan)
		if !ok {
			showError(c, nil, "Ooops...", "internal processing error")
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
					showError(c, err, "Response processing failed", "")
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
			showError(c, nil, "Upload failed", "No file uploaded or invalid file")
			return
		}

		// Verify the file is an EPUB
		if filepath.Ext(file.Filename) != ".epub" {
			pkg.Logger.Printf("Invalid file type: %s\n", file.Filename)
			showError(c, nil, "Bad format", "only EPUB files are allowed")
			return
		}

		ff, err := file.Open()
		defer ff.Close()
		if err != nil {
			pkg.Logger.Printf("Error opening file: %v\n", err)
			showError(c, nil, "Upload failed", "Failed to open file")
		}

		content, err := io.ReadAll(ff)
		if err != nil {
			pkg.Logger.Printf("Error reading file: %v\n", err)
			showError(c, nil, "Upload failed", "Failed to read file")
			return
		}

		timestamp := time.Now().Unix()
		uniqueFilename := fmt.Sprintf("%d_%s", timestamp, file.Filename)
		f := domain.File{Name: uniqueFilename}
		fc := &domain.FileWithContent{
			File:    f,
			Content: content,
		}

		// Parse the EPUB file
		book, err := s.bookRepository.ReadFromFile(c.Request.Context(), fc)
		if err != nil {
			pkg.Logger.Printf("Error processing EPUB: %v", err)
			showError(c, nil, "Upload failed", "Failed to process EPUB file. Is your file corrupted?")
			return
		}

		// Check if the book already exists
		if _, err := s.bookRepository.GetByTitleAndAuthor(c.Request.Context(), book.Title, book.Author); err == nil {
			if errors.Is(domain.ErrBookAlreadyExists, err) {
				showInfo(c, "Book already exists",
					"A book with the same title and author already exists in the library")
				return
			}
			pkg.Logger.Printf("Error getting book: %v", err)
			showError(c, nil, "Upload failed", "Failed to get book")
			return
		}

		if err := s.fileRepository.Store(c.Request.Context(), fc); err != nil {
			pkg.Logger.Printf("Error storing file: %v\n", err)
			showError(c, nil, "Upload failed", "Failed to store file")
			return
		}

		// Add the book to the repository
		addedBook, err := s.bookRepository.Add(
			c.Request.Context(),
			book.Title,
			book.Author,
			f,
			book.Metadata,
			domain.WithStatus(domain.StatusIndexing),
		)
		if err != nil {
			pkg.Logger.Printf("Error adding book to repository: %v", err)
			showError(c, nil, "Upload failed", "Failed to add book to library")
			return
		}

		showSuccess(c, "Book uploaded",
			fmt.Sprintf("Book %s uploaded successfully. Indexing is starting...", addedBook.Title))
		c.HTML(http.StatusOK, "", components.BookCard(addedBook))
	})
}
