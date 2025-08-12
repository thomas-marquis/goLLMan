package domain

import (
	"context"

	"github.com/firebase/genkit/go/ai"
)

type BookVectorStore interface {
	Index(ctx context.Context, book Book, contents []string, vectors [][]float32) error
	Retrieve(ctx context.Context, books []Book, embedding []float32, limit int) ([]*ai.Document, error)
}
