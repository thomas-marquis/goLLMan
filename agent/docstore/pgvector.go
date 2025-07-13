package docstore

import (
	"context"
	"errors"
	"fmt"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
	"github.com/thomas-marquis/goLLMan/internal/domain"
	"github.com/thomas-marquis/goLLMan/pkg"
	"github.com/tmc/langchaingo/textsplitter"
)

const (
	queryInsertBookIndex = "INSERT INTO book_index (book_id, content, embedding) VALUES ($1, $2, $3)"
)

type PgVectorStore struct {
	pool       *pgxpool.Pool
	retriever  ai.Retriever
	splitter   textsplitter.TextSplitter
	embedder   ai.Embedder
	repository domain.BookRepository
}

var _ DocStore = (*PgVectorStore)(nil)

func NewPgVectorStore(
	g *genkit.Genkit,
	pool *pgxpool.Pool,
	repository domain.BookRepository,
	embeddingModel string,
) (*PgVectorStore, error) {
	pkg.Logger.Printf("Initializing pgvector store with embedding model: %s\n", embeddingModel)

	embedder, err := getEmbedder(g, embeddingModel)
	if err != nil {
		return nil, err
	}

	splitter := makeTextSplitter()

	return &PgVectorStore{
		pool:       pool,
		splitter:   splitter,
		embedder:   embedder,
		repository: repository,
	}, nil
}

func (s *PgVectorStore) Index(ctx context.Context, book domain.Book, docs []*ai.Document) error {
	var bookID int
	var err error
	_, bookID, err = s.getBookByTitleAndAuthor(ctx, book.Title, book.Author)
	if err != nil {
		if !errors.Is(err, errBookDoesntExists) {
			bookID, err = s.saveBook(ctx, book)
			if err != nil {
				return err
			}
		}
		return err
	}

	preparedDocs, err := splitDocuments(s.splitter, docs)
	if err != nil {
		return fmt.Errorf("failed to split documents: %w", err)
	}

	documentsBatches := batchDocuments(preparedDocs, 50)
	pkg.Logger.Printf("Indexing %d documents for book %s through %d batches",
		len(preparedDocs), book.Title, len(documentsBatches))

	const nbValues = 3
	for i, batch := range documentsBatches {
		batchSize := len(batch)
		if batchSize == 0 {
			continue
		}

		embedRes, err := ai.Embed(ctx, s.embedder, ai.WithDocs(batch...))
		if err != nil {
			return fmt.Errorf("failed to embed documents: %w", err)
		}

		batchValues := make([]any, 0, batchSize*3)
		for i, doc := range batch {
			content := pkg.ContentToText(doc.Content)
			embedding := pgvector.NewVector(embedRes.Embeddings[i].Embedding).String()
			batchValues = append(batchValues, bookID, content, embedding)
		}

		tx, err := s.pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		for j := 0; j < batchSize; j++ {
			vals := batchValues[j*nbValues : j*nbValues+nbValues]
			if _, err := tx.Exec(ctx, queryInsertBookIndex, vals...); err != nil {
				tx.Rollback(ctx)
				return fmt.Errorf("failed to insert document with values %v: %w", vals, err)
			}
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit transaction for batch %d: %w", i, err)
		}

		pkg.Logger.Printf("Batch no %d/%d indexed with %d documents for book %s",
			i+1, len(documentsBatches), batchSize, book.Title)
	}
	return nil
}

func (s *PgVectorStore) Retrieve(ctx context.Context, book domain.Book, query string, limit int) ([]*ai.Document, error) {
	//TODO implement me
	panic("implement me")
}
