package loader

import (
	"fmt"
	"github.com/firebase/genkit/go/ai"
	"github.com/thomas-marquis/goLLMan/internal/domain"
	"github.com/timsims/pamphlet"
)

type LocalEpubLoader struct {
	repository domain.BookRepository
}

var _ BookLoader = (*LocalEpubLoader)(nil)

func NewLocalEpubLoader(repository domain.BookRepository) *LocalEpubLoader {
	return &LocalEpubLoader{repository}
}

func (l *LocalEpubLoader) Load(book domain.Book) ([]*ai.Document, error) {
	filePath, ok := book.Metadata["local_epub_filepath"].(string)
	if !ok || filePath == "" {
		return nil, fmt.Errorf("book %s by %s does not have a valid local epub filepath",
			book.Title, book.Author)
	}

	parser, err := pamphlet.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening pdf at %s: %w", filePath, err)
	}

	parsedBook := parser.GetBook()

	documents := make([]*ai.Document, len(parsedBook.Chapters), len(parsedBook.Chapters))
	chapters := parsedBook.Chapters
	for i, chapter := range chapters {
		chapContent, err := chapter.GetContent()
		if err != nil {
			return nil, fmt.Errorf("failed to get content for chapter %d: %w", i+1, err)
		}
		documents[i] = ai.DocumentFromText(chapContent, map[string]any{
			"title":   chapter.Title,
			"book_id": book.ID,
		})
	}

	return documents, nil
}
