package docstore

import (
	"context"
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
	queryInsertBookIndex       = "INSERT INTO book_index (book_id, content, embedding) VALUES ($1, $2, $3)"
	querySearchNearestByBookId = `
SELECT content
FROM book_index bi
WHERE book_id = $1
ORDER BY embedding <=> $2
LIMIT $3`
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

	s := &PgVectorStore{
		pool:       pool,
		splitter:   splitter,
		embedder:   embedder,
		repository: repository,
	}

	genkit.DefineRetriever(g, "gollman", "pgvector",
		func(ctx context.Context, req *ai.RetrieverRequest) (*ai.RetrieverResponse, error) {
			book_id, ok := req.Query.Metadata["book_id"].(string)
			if !ok {
				return nil, fmt.Errorf("book not found in query metadata")
			}
			limit, ok := req.Query.Metadata["limit"].(int)
			if !ok || limit <= 0 {
				limit = 3
				pkg.Logger.Printf("limit not found in query metadata, applying default limit: %d", limit)
			}
			query := pkg.ContentToText(req.Query.Content)

			book, err := s.repository.GetByID(ctx, book_id)
			if err != nil {
				return nil, fmt.Errorf("failed to retrieve book by ID %s: %w", book_id, err)
			}

			docs, err := s.Retrieve(ctx, book, query, limit)
			if err != nil {
				return nil, fmt.Errorf("failed to retrieve documents: %w", err)
			}

			return &ai.RetrieverResponse{Documents: docs}, nil
		})

	return s, nil
}

func (s *PgVectorStore) Index(ctx context.Context, book domain.Book, docs []*ai.Document) error {
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

		batchValues := make([]any, 0, batchSize*nbValues)
		for i, doc := range batch {
			content := pkg.ContentToText(doc.Content)
			embedding := pgvector.NewVector(embedRes.Embeddings[i].Embedding).String()
			batchValues = append(batchValues, book.ID, content, embedding)
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
	eReq := &ai.EmbedRequest{
		Input: []*ai.Document{ai.DocumentFromText(query, nil)},
	}
	eRes, err := s.embedder.Embed(ctx, eReq)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}
	vec := eRes.Embeddings[0].Embedding

	rows, err := s.pool.Query(ctx, querySearchNearestByBookId, book.ID, pgvector.NewVector(vec), limit)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search query: %w", err)
	}

	documents := make([]*ai.Document, 0, limit)
	for rows.Next() {
		var content string
		if err := rows.Scan(&content); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		doc := ai.DocumentFromText(content, map[string]any{"book_id": book.ID})
		documents = append(documents, doc)
	}

	return documents, nil
}
