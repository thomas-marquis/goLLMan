package loader

import (
	"github.com/firebase/genkit/go/ai"
	"github.com/thomas-marquis/goLLMan/internal/domain"
)

type BookLoader interface {
	Load(book domain.Book) ([]*ai.Document, error)
}
