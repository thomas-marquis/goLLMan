package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/timsims/pamphlet"
	"os"
	"strconv"
	"sync"

	"github.com/thomas-marquis/goLLMan/internal/domain"
)

const (
	bookMetaLocalEpubPathKey = "local_epub_filepath"
)

// BookRepositoryLocal represents a local repository of books.
type BookRepositoryLocal struct {
	filePath string
	mu       sync.Mutex
}

var _ domain.BookRepository = (*BookRepositoryLocal)(nil)

// NewBookRepositoryLocal creates a new instance of BookRepositoryLocal.
func NewBookRepositoryLocal(jsonFile string) *BookRepositoryLocal {
	return &BookRepositoryLocal{
		filePath: jsonFile,
	}
}

// List retrieves the list of books.
func (r *BookRepositoryLocal) List(ctx context.Context) ([]domain.Book, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	data, err := os.ReadFile(r.filePath)
	if err != nil {
		return nil, errors.Join(domain.ErrRepositoryError, err)
	}

	var books []domain.Book
	if err := json.Unmarshal(data, &books); err != nil {
		return nil, errors.Join(domain.ErrRepositoryError, err)
	}

	return books, nil
}

// Add adds a new book to the repository.
func (r *BookRepositoryLocal) Add(ctx context.Context, title, author string, metadata map[string]any) (domain.Book, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	books, err := r.readBooksFromFile()
	if err != nil {
		return domain.Book{}, errors.Join(domain.ErrRepositoryError, err)
	}

	// Check if the book already exists
	for _, book := range books {
		if book.Title == title && book.Author == author {
			return book, nil
		}
	}

	// If not, create a new book
	// Generate a simple ID for the new book
	id := generateID(books)
	newBook := domain.Book{
		ID:       id,
		Title:    title,
		Author:   author,
		Metadata: metadata,
	}

	books = append(books, newBook)

	if err := r.writeBooksToFile(books); err != nil {
		return domain.Book{}, errors.Join(domain.ErrRepositoryError, err)
	}

	return newBook, nil
}

// GetByID retrieves a book by its ID.
func (r *BookRepositoryLocal) GetByID(ctx context.Context, id string) (domain.Book, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	books, err := r.readBooksFromFile()
	if err != nil {
		return domain.Book{}, errors.Join(domain.ErrRepositoryError, err)
	}

	for _, book := range books {
		if book.ID == id {
			return book, nil
		}
	}

	return domain.Book{}, domain.ErrBookNotFound
}

// GetByTitleAndAuthor retrieves a book by its title and author.
func (r *BookRepositoryLocal) GetByTitleAndAuthor(ctx context.Context, title, author string) (domain.Book, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	books, err := r.readBooksFromFile()
	if err != nil {
		return domain.Book{}, errors.Join(domain.ErrRepositoryError, err)
	}

	for _, book := range books {
		if book.Title == title && book.Author == author {
			return book, nil
		}
	}

	return domain.Book{}, domain.ErrBookNotFound
}

func (r *BookRepositoryLocal) ReadFromFile(ctx context.Context, filePath string) (domain.Book, error) {
	return parseEpubFromFile(filePath)
}

// readBooksFromFile reads the list of books from the file.
func (r *BookRepositoryLocal) readBooksFromFile() ([]domain.Book, error) {
	data, err := os.ReadFile(r.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []domain.Book{}, nil
		}
		return nil, errors.Join(domain.ErrRepositoryError, err)
	}

	var books []domain.Book
	if err := json.Unmarshal(data, &books); err != nil {
		return nil, errors.Join(domain.ErrRepositoryError, err)
	}

	return books, nil
}

// writeBooksToFile writes the list of books to the file.
func (r *BookRepositoryLocal) writeBooksToFile(books []domain.Book) error {
	data, err := json.Marshal(books)
	if err != nil {
		return err
	}

	return os.WriteFile(r.filePath, data, 0644)
}

// generateID generates a simple ID for a new book.
func generateID(books []domain.Book) string {
	maxID := 0
	for _, book := range books {
		id, err := strconv.Atoi(book.ID)
		if err == nil && id > maxID {
			maxID = id
		}
	}
	return fmt.Sprintf("%d", maxID+1)
}

func parseEpubFromFile(filePath string) (domain.Book, error) {
	parser, err := pamphlet.Open(filePath)
	if err != nil {
		return domain.Book{}, fmt.Errorf("error opening epub at %s: %w", filePath, err)
	}

	book := parser.GetBook()
	if book == nil {
		return domain.Book{}, fmt.Errorf("no book found in epub at %s", filePath)
	}

	return domain.Book{
		Title:  book.Title,
		Author: book.Author,
		Metadata: map[string]any{
			bookMetaLocalEpubPathKey: filePath,
		},
	}, nil
}
