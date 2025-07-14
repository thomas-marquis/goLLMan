package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thomas-marquis/goLLMan/internal/domain"
	"strconv"
)

const (
	querySelectBookByID             = "SELECT id, title, author, metadata FROM books WHERE id = $1"
	queryInsertBook                 = "INSERT INTO books (title, author, metadata) VALUES ($1, $2, $3) RETURNING id"
	querySelectAllBooks             = "SELECT id, title, author, metadata FROM books"
	querySelectBookByTitleAndAuthor = "SELECT id, title, author, metadata FROM books WHERE title = $1 AND author = $2"
)

type BookRepositoryPostgres struct {
	pool *pgxpool.Pool
}

var _ domain.BookRepository = (*BookRepositoryPostgres)(nil)

func NewBookRepositoryPostgres(pool *pgxpool.Pool) *BookRepositoryPostgres {
	return &BookRepositoryPostgres{
		pool: pool,
	}
}

func (r *BookRepositoryPostgres) List(ctx context.Context) ([]domain.Book, error) {
	rows, err := r.pool.Query(ctx, querySelectAllBooks)
	if err != nil {
		return nil, errors.Join(domain.ErrRepositoryError, err)
	}
	defer rows.Close()

	var books []domain.Book
	for rows.Next() {
		var book domain.Book
		var metadataJson []byte

		if err := rows.Scan(&book.ID, &book.Title, &book.Author, &metadataJson); err != nil {
			return nil, errors.Join(domain.ErrRepositoryError, err)
		}

		if metadataJson != nil {
			var metadata map[string]any
			if err := json.Unmarshal(metadataJson, &metadata); err != nil {
				return nil, errors.Join(domain.ErrRepositoryError, err)
			}
			book.Metadata = metadata
		} else {
			book.Metadata = nil
		}
		books = append(books, book)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Join(domain.ErrRepositoryError, err)
	}

	return books, nil
}

func (r *BookRepositoryPostgres) Add(ctx context.Context, title, author string, metadata map[string]any) (domain.Book, error) {
	// Check if the book already exists
	var book domain.Book
	var err error

	exists := true
	book, err = r.GetByTitleAndAuthor(ctx, title, author)
	if err != nil {
		if errors.Is(err, domain.ErrBookNotFound) {
			exists = false
		} else {
			return domain.Book{}, errors.Join(domain.ErrRepositoryError, err)
		}
	}
	if exists {
		return book, nil
	}

	// If the book does not exist, insert it
	var metadataJson []byte

	if metadata != nil {
		metadataJson, err = json.Marshal(metadata)
		if err != nil {
			return domain.Book{}, errors.Join(domain.ErrRepositoryError, err)
		}
	}

	if err := r.pool.QueryRow(ctx, queryInsertBook, title, author, metadataJson).Scan(&book.ID); err != nil {
		return domain.Book{}, errors.Join(domain.ErrRepositoryError, err)
	}
	book.Title = title
	book.Author = author

	return book, nil
}

func (r *BookRepositoryPostgres) GetByID(ctx context.Context, id string) (domain.Book, error) {
	var book domain.Book
	var metadataJson []byte

	bookId, err := strconv.Atoi(id)
	if err != nil {
		return book, errors.Join(domain.ErrRepositoryError, err)
	}
	if err := r.pool.QueryRow(ctx, querySelectBookByID, bookId).
		Scan(&book.ID, &book.Title, &book.Author, &metadataJson); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return book, domain.ErrBookNotFound
		}
		return book, errors.Join(domain.ErrRepositoryError, err)
	}

	if metadataJson != nil {
		var metadata map[string]any
		if err := json.Unmarshal(metadataJson, &metadata); err != nil {
			return book, errors.Join(domain.ErrRepositoryError, err)
		}
		book.Metadata = metadata
	} else {
		book.Metadata = nil
	}

	return book, nil
}

func (r *BookRepositoryPostgres) GetByTitleAndAuthor(ctx context.Context, title, author string) (domain.Book, error) {
	var book domain.Book
	var metadataJson []byte

	if err := r.pool.QueryRow(ctx, querySelectBookByTitleAndAuthor, title, author).
		Scan(&book.ID, &book.Title, &book.Author, &metadataJson); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return book, domain.ErrBookNotFound
		}
		return book, errors.Join(domain.ErrRepositoryError, err)
	}

	if metadataJson != nil {
		var metadata map[string]any
		if err := json.Unmarshal(metadataJson, &metadata); err != nil {
			return book, errors.Join(domain.ErrRepositoryError, err)
		}
		book.Metadata = metadata
	} else {
		book.Metadata = nil
	}

	return book, nil
}

func (r *BookRepositoryPostgres) ReadFromFile(ctx context.Context, filePath string) (domain.Book, error) {
	return parseEpubFromFile(filePath)
}
