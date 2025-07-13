package infrastructure

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thomas-marquis/goLLMan/internal/domain"
	"strconv"
)

const (
	querySelectBookByID             = "SELECT id, title, author FROM books WHERE id = $1"
	queryInsertBook                 = "INSERT INTO books (title, author) VALUES ($1, $2) RETURNING id"
	querySelectAllBooks             = "SELECT id, title, author FROM books"
	querySelectBookByTitleAndAuthor = "SELECT id, title, author FROM books WHERE title = $1 AND author = $2"
)

type BookRepositoryPostgres struct {
	pool *pgxpool.Pool
}

func NewBookRepository(pool *pgxpool.Pool) *BookRepositoryPostgres {
	return &BookRepositoryPostgres{
		pool: pool,
	}
}

func (b *BookRepositoryPostgres) List(ctx context.Context) ([]domain.Book, error) {
	rows, err := b.pool.Query(ctx, querySelectAllBooks)
	if err != nil {
		return nil, errors.Join(domain.ErrRepositoryError, err)
	}
	defer rows.Close()

	var books []domain.Book
	for rows.Next() {
		var book domain.Book
		if err := rows.Scan(&book.ID, &book.Title, &book.Author); err != nil {
			return nil, errors.Join(domain.ErrRepositoryError, err)
		}
		books = append(books, book)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Join(domain.ErrRepositoryError, err)
	}

	return books, nil
}

func (b *BookRepositoryPostgres) Add(ctx context.Context, title, author string) (domain.Book, error) {
	// Check if the book already exists
	var book domain.Book
	var err error

	exists := true
	book, err = b.GetByTitleAndAuthor(ctx, title, author)
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
	if err := b.pool.QueryRow(ctx, queryInsertBook, title, author).Scan(&book.ID); err != nil {
		return domain.Book{}, errors.Join(domain.ErrRepositoryError, err)
	}
	book.Title = title
	book.Author = author

	return book, nil
}

func (b *BookRepositoryPostgres) GetByID(ctx context.Context, id string) (domain.Book, error) {
	var book domain.Book

	bookId, err := strconv.Atoi(id)
	if err != nil {
		return book, errors.Join(domain.ErrRepositoryError, err)
	}
	if err := b.pool.QueryRow(ctx, querySelectBookByID, bookId).Scan(&book.ID, &book.Title, &book.Author); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return book, domain.ErrBookNotFound
		}
		return book, errors.Join(domain.ErrRepositoryError, err)
	}
	return book, nil
}

func (b *BookRepositoryPostgres) GetByTitleAndAuthor(ctx context.Context, title, author string) (domain.Book, error) {
	var book domain.Book
	err := b.pool.QueryRow(ctx, querySelectBookByTitleAndAuthor, title, author).Scan(&book.ID, &book.Title, &book.Author)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return book, domain.ErrBookNotFound
		}
		return book, errors.Join(domain.ErrRepositoryError, err)
	}

	return book, nil
}
