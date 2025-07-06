package docstore

import (
	"context"
	"github.com/firebase/genkit/go/ai"
)

type DocStore interface {
	Index(ctx context.Context, docs []*ai.Document) error
	Retrieve(ctx context.Context, query string, limit int) ([]*ai.Document, error)
}
