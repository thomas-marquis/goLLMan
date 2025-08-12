package agent

import (
	"context"
	"fmt"

	"github.com/firebase/genkit/go/ai"
	"github.com/thomas-marquis/goLLMan/internal/domain"
	"github.com/thomas-marquis/goLLMan/pkg"
)

const (
	indexingBatchSize = 50
)

type dataToInsert struct {
	Content   string
	Embedding []float32
	Index     int
	Batch     int
}

const numInsertWorkers = 10

func indexDocuments(embedder ai.Embedder, bookVectorStore domain.BookVectorStore, ctx context.Context, book domain.Book, docs []*ai.Document) error {
	splitter := makeTextSplitter()

	preparedDocs, err := SplitDocuments(splitter, docs)
	if err != nil {
		return fmt.Errorf("failed to split documents: %w", err)
	}

	documentsBatches := batchDocuments(preparedDocs, indexingBatchSize)
	pkg.Logger.Printf("Indexing %d documents for book %s through %d batches",
		len(preparedDocs), book.Title, len(documentsBatches))

	for i, batch := range documentsBatches {
		batchSize := len(batch)
		if batchSize == 0 {
			continue
		}

		embedRes, err := ai.Embed(ctx, embedder, ai.WithDocs(batch...))
		if err != nil {
			return fmt.Errorf("failed to embed documents: %w", err)
		}

		vectors := make([][]float32, batchSize)
		for j, emb := range embedRes.Embeddings {
			vectors[j] = emb.Embedding
		}

		contents := make([]string, batchSize)
		for j, doc := range batch {
			contents[j] = pkg.ContentToText(doc.Content)
		}

		if err := bookVectorStore.Index(ctx, book, contents, vectors); err != nil {
			return fmt.Errorf("failed to index documents at batch %d: %w", i, err)
		}

		pkg.Logger.Printf("Batch no %d/%d indexed with %d documents for book %s",
			i+1, len(documentsBatches), batchSize, book.Title)
	}

	return nil
}
