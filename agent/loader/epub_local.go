package loader

import (
	"fmt"
	"github.com/firebase/genkit/go/ai"
	"github.com/thomas-marquis/goLLMan/internal/domain"
	"github.com/timsims/pamphlet"
)

type chapterContent struct {
	Title   string
	Content string
}

type LocalEpubLoader struct {
	filePath   string
	repository domain.BookRepository
}

var _ BookLoader = (*LocalEpubLoader)(nil)

func NewLocalEpubLoader(path string, repository domain.BookRepository) *LocalEpubLoader {
	return &LocalEpubLoader{path, repository}
}

func (l *LocalEpubLoader) List() ([]domain.Book, error) {
	parser, err := pamphlet.Open(l.filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening epub at %s: %w", l.filePath, err)
	}

	book := parser.GetBook()
	if book == nil {
		return nil, fmt.Errorf("no book found in epub at %s", l.filePath)
	}

	return []domain.Book{{
		Title:  book.Title,
		Author: book.Author,
	}}, nil
}

func (l *LocalEpubLoader) Load(bookId string) (domain.Book, []*ai.Document, error) {
	parser, err := pamphlet.Open(l.filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening pdf at %s: %w", l.filePath, err)
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
			"title": chapter.Title,
		})
	}

	return documents, nil
}
