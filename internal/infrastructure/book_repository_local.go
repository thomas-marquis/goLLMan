package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/thomas-marquis/goLLMan/internal/domain"
)

// BookRepositoryLocal represents a local repository of books.
type BookRepositoryLocal struct {
	filePath string
	mu       sync.Mutex
}

// NewBookRepositoryLocal creates a new instance of BookRepositoryLocal.
func NewBookRepositoryLocal(jsonFile string) *BookRepositoryLocal {
	return &BookRepositoryLocal{
		filePath: jsonFile,
	}
}

// List retrieves the list of books.
func (b *BookRepositoryLocal) List(ctx context.Context) ([]domain.Book, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	data, err := os.ReadFile(b.filePath)
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
func (b *BookRepositoryLocal) Add(ctx context.Context, title, author string) (domain.Book, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	books, err := b.readBooksFromFile()
	if err != nil {
		return domain.Book{}, errors.Join(domain.ErrRepositoryError, err)
	}

	// Generate a simple ID for the new book
	id := generateID(books)
	newBook := domain.Book{
		ID:     id,
		Title:  title,
		Author: author,
	}

	books = append(books, newBook)

	if err := b.writeBooksToFile(books); err != nil {
		return domain.Book{}, errors.Join(domain.ErrRepositoryError, err)
	}

	return newBook, nil
}

// GetByID retrieves a book by its ID.
func (b *BookRepositoryLocal) GetByID(ctx context.Context, id string) (domain.Book, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	books, err := b.readBooksFromFile()
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
func (b *BookRepositoryLocal) GetByTitleAndAuthor(ctx context.Context, title, author string) (domain.Book, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	books, err := b.readBooksFromFile()
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

// readBooksFromFile reads the list of books from the file.
func (b *BookRepositoryLocal) readBooksFromFile() ([]domain.Book, error) {
	data, err := os.ReadFile(b.filePath)
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
func (b *BookRepositoryLocal) writeBooksToFile(books []domain.Book) error {
	data, err := json.Marshal(books)
	if err != nil {
		return err
	}

	return os.WriteFile(b.filePath, data, 0644)
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
