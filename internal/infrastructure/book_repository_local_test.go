package infrastructure_test

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/thomas-marquis/goLLMan/internal/domain"
	"github.com/thomas-marquis/goLLMan/internal/infrastructure"
	"os"
	"path/filepath"
	"testing"
)

func TestBookRepositoryLocal_List(t *testing.T) {
	// Given
	dir := t.TempDir()
	filePath := filepath.Join(dir, "test_books.json")

	initialBooks := []domain.Book{
		{ID: "1", Title: "Book 1", Author: "Author 1"},
		{ID: "2", Title: "Book 2", Author: "Author 2"},
	}
	writeBooksToFile(t, filePath, initialBooks)

	repo := infrastructure.NewBookRepositoryLocal(filePath)

	// When
	books, err := repo.List(context.TODO())

	// Then
	assert.NoError(t, err)
	assert.Equal(t, initialBooks, books)
}

func TestBookRepositoryLocal_Add(t *testing.T) {
	// Given
	dir := t.TempDir()
	filePath := filepath.Join(dir, "test_books.json")

	repo := infrastructure.NewBookRepositoryLocal(filePath)

	// When
	newBook, err := repo.Add(context.TODO(), "New Book", "New Author", nil)

	// Then
	assert.NoError(t, err)
	assert.NotEmpty(t, newBook.ID)
	assert.Equal(t, "1", newBook.ID)

	books, err := repo.List(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, 1, len(books))
	assert.Equal(t, "New Book", books[0].Title)
	assert.Equal(t, "New Author", books[0].Author)
}

func TestBookRepositoryLocal_Add_ShouldIncrementId(t *testing.T) {
	// Given
	dir := t.TempDir()
	filePath := filepath.Join(dir, "test_books.json")

	initialBooks := []domain.Book{
		{ID: "1", Title: "Book 1", Author: "Author 1"},
		{ID: "2", Title: "Book 2", Author: "Author 2"},
	}
	writeBooksToFile(t, filePath, initialBooks)

	repo := infrastructure.NewBookRepositoryLocal(filePath)

	// When
	newBook, err := repo.Add(context.TODO(), "New Book", "New Author", nil)

	// Then
	assert.NoError(t, err)
	assert.NotEmpty(t, newBook.ID)
	assert.Equal(t, "3", newBook.ID)

	books, err := repo.List(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, 3, len(books))
	assert.Equal(t, "New Book", books[2].Title)
	assert.Equal(t, "New Author", books[2].Author)
}

func TestBookRepositoryLocal_GetByID_WhenBookExists(t *testing.T) {
	// Given
	dir := t.TempDir()
	filePath := filepath.Join(dir, "test_books.json")

	initialBooks := []domain.Book{
		{ID: "1", Title: "Book 1", Author: "Author 1"},
		{ID: "2", Title: "Book 2", Author: "Author 2"},
	}
	writeBooksToFile(t, filePath, initialBooks)

	repo := infrastructure.NewBookRepositoryLocal(filePath)

	// When
	book, err := repo.GetByID(context.TODO(), "2")

	// Then
	assert.NoError(t, err)
	assert.Equal(t, "Book 2", book.Title)
	assert.Equal(t, "Author 2", book.Author)

	// Test non-existent ID
	_, err = repo.GetByID(context.TODO(), "99")
	assert.Error(t, err)
}

func TestBookRepositoryLocal_GetByID_WhenBookDoesNotExist(t *testing.T) {
	// Given
	dir := t.TempDir()
	filePath := filepath.Join(dir, "test_books.json")

	initialBooks := []domain.Book{
		{ID: "1", Title: "Book 1", Author: "Author 1"},
	}
	writeBooksToFile(t, filePath, initialBooks)

	repo := infrastructure.NewBookRepositoryLocal(filePath)

	// When
	_, err := repo.GetByID(context.TODO(), "99")

	// Then
	assert.Error(t, err)
	assert.Equal(t, domain.ErrBookNotFound, err)
}

func TestBookRepositoryLocal_GetByTitleAndAuthor_WhenBookExists(t *testing.T) {
	// Given
	dir := t.TempDir()
	filePath := filepath.Join(dir, "test_books.json")

	initialBooks := []domain.Book{
		{ID: "1", Title: "Book 1", Author: "Author 1"},
		{ID: "2", Title: "Book 2", Author: "Author 2"},
	}
	writeBooksToFile(t, filePath, initialBooks)

	repo := infrastructure.NewBookRepositoryLocal(filePath)

	// When
	book, err := repo.GetByTitleAndAuthor(context.TODO(), "Book 2", "Author 2")

	// Then
	assert.NoError(t, err)
	assert.Equal(t, "2", book.ID)

	// Test non-existent book
	_, err = repo.GetByTitleAndAuthor(context.TODO(), "Nonexistent", "Author")
	assert.Error(t, err)
}

func TestBookRepositoryLocal_GetByTitleAndAuthor_WhenBookDoesNotExist(t *testing.T) {
	// Given
	dir := t.TempDir()
	filePath := filepath.Join(dir, "test_books.json")

	initialBooks := []domain.Book{
		{ID: "1", Title: "Book 1", Author: "Author 1"},
	}
	writeBooksToFile(t, filePath, initialBooks)

	repo := infrastructure.NewBookRepositoryLocal(filePath)

	// When
	_, err := repo.GetByTitleAndAuthor(context.TODO(), "Nonexistent", "Author")
	assert.Error(t, err)
	assert.Equal(t, domain.ErrBookNotFound, err)
}

// Helper function to write books to a file
func writeBooksToFile(t *testing.T, filePath string, books []domain.Book) {
	t.Helper()
	data, _ := json.Marshal(books)
	os.WriteFile(filePath, data, 0644)
}
