package infrastructure_test

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/thomas-marquis/goLLMan/internal/infrastructure"
	"testing"
	"time"
)

func Test_BookRepositoryPostgres(t *testing.T) {
	// Use testcontainers to launch a PostgreSQL instance
	ctx := context.Background()
	postgresContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("docker.io/postgres:17"),
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, postgresContainer.Terminate(ctx))
	}()

	connStr, err := postgresContainer.ConnectionString(ctx)
	assert.NoError(t, err)

	pool, err := pgxpool.New(ctx, connStr)
	assert.NoError(t, err)
	defer pool.Close()

	// Migrate the schema
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS books (
			id SERIAL PRIMARY KEY,
			title TEXT NOT NULL,
			author TEXT NOT NULL,
			metadata JSONB
		);
	`)
	assert.NoError(t, err)

	repo := infrastructure.NewBookRepositoryPostgres(pool)

	t.Run("Add and List Books", func(t *testing.T) {
		metadata := map[string]interface{}{
			"genre":               "Programming",
			"year":                2021,
			"local_epub_filepath": "path/to/book.epub",
		}

		book, err := repo.Add(ctx, "The Go Programming Language", "Alan A. A. Donovan and Brian W. Kernighan", metadata)
		assert.NoError(t, err)
		assert.NotZero(t, book.ID)

		books, err := repo.List(ctx)
		assert.NoError(t, err)
		assert.Len(t, books, 1)
		assert.Equal(t, "The Go Programming Language", books[0].Title)
	})

	t.Run("Get Book By ID", func(t *testing.T) {
		book, err := repo.GetByID(ctx, "1")
		assert.NoError(t, err)
		assert.Equal(t, "The Go Programming Language", book.Title)
	})

	t.Run("Get Book By Title and Author", func(t *testing.T) {
		book, err := repo.GetByTitleAndAuthor(ctx, "The Go Programming Language", "Alan A. A. Donovan and Brian W. Kernighan")
		assert.NoError(t, err)
		assert.Equal(t, "The Go Programming Language", book.Title)
	})
}
