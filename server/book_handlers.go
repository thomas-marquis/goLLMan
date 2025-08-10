package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/thomas-marquis/goLLMan/controller/server/components"
	"github.com/thomas-marquis/goLLMan/pkg"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// UploadBookHandler handles the EPUB file upload
func (s *Server) UploadBookHandler(router *gin.Engine) {
	router.POST("/books/upload", func(c *gin.Context) {
		// Get the file from the request
		file, err := c.FormFile("epub-file")
		if err != nil {
			components.ErrorMessage("No file uploaded or invalid file").Render(c.Request.Context(), c.Writer)
			return
		}

		// Verify the file is an EPUB
		if filepath.Ext(file.Filename) != ".epub" {
			components.ErrorMessage("Only EPUB files are allowed").Render(c.Request.Context(), c.Writer)
			return
		}

		// Create temporary directory if it doesn't exist
		tempDir := "./tmp/uploads"
		if err := os.MkdirAll(tempDir, 0755); err != nil {
			pkg.Logger.Printf("Error creating directory: %v", err)
			components.ErrorMessage("Failed to prepare for upload").Render(c.Request.Context(), c.Writer)
			return
		}

		// Generate a unique filename to avoid collisions
		timestamp := time.Now().Unix()
		tempFilePath := fmt.Sprintf("%s/%d_%s", tempDir, timestamp, file.Filename)

		// Save the file to disk
		if err := c.SaveUploadedFile(file, tempFilePath); err != nil {
			pkg.Logger.Printf("Error saving file: %v", err)
			components.ErrorMessage("Failed to save file").Render(c.Request.Context(), c.Writer)
			return
		}

		// Process the EPUB file
		book, err := s.bookRepository.ReadFromFile(c.Request.Context(), tempFilePath)
		if err != nil {
			pkg.Logger.Printf("Error processing EPUB: %v", err)
			// Clean up the temp file
			os.Remove(tempFilePath)
			components.ErrorMessage("Failed to process EPUB file").Render(c.Request.Context(), c.Writer)
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
			pkg.Logger.Printf("Error adding book to repository: %v", err)
			// Clean up the temp file
			os.Remove(tempFilePath)
			components.ErrorMessage("Failed to add book to library").Render(c.Request.Context(), c.Writer)
			return
		}

		// Clean up the temp file
		os.Remove(tempFilePath)

		// Return success with HTML
		components.SuccessMessage("Book uploaded successfully", addedBook).Render(c.Request.Context(), c.Writer)
	})
}
