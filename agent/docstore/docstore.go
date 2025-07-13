package docstore

import (
	"context"
	"github.com/firebase/genkit/go/ai"
	"github.com/thomas-marquis/goLLMan/internal/domain"
)

const (
	DocStoreTypeLocal    = "local"
	DocStoreTypePgvector = "pgvector"
)

type DocStore interface {
	Index(ctx context.Context, book domain.Book, docs []*ai.Document) error
	Retrieve(ctx context.Context, book domain.Book, query string, limit int) ([]*ai.Document, error)
}
