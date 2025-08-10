package domain

import (
	"context"
	"github.com/firebase/genkit/go/ai"
)

type BookVectorStore interface {
	Index(ctx context.Context, book Book, content string, vector []float32) error
	Retrieve(ctx context.Context, book Book, embedding []float32, limit int) ([]*ai.Document, error)
}
