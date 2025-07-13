package loader

import (
	"github.com/firebase/genkit/go/ai"
	"github.com/thomas-marquis/goLLMan/internal/domain"
)

type BookLoader interface {
	List() ([]domain.Book, error)
	Load(bookId string) (domain.Book, []*ai.Document, error)
}
