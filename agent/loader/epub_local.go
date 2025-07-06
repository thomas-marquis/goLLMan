package loader

import (
	"fmt"
	"github.com/firebase/genkit/go/ai"
	"github.com/timsims/pamphlet"
)

type chapterContent struct {
	Title   string
	Content string
}

type LocalEpubLoader struct {
	filePath string
}

func NewLocalEpubLoader(path string) *LocalEpubLoader {
	return &LocalEpubLoader{path}
}

func (l *LocalEpubLoader) Load() ([]*ai.Document, error) {
	parser, err := pamphlet.Open(l.filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening pdf at %s: %w", l.filePath, err)
	}

	book := parser.GetBook()

	documents := make([]*ai.Document, len(book.Chapters), len(book.Chapters))
	chapters := book.Chapters
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
