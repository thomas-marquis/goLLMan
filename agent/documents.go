package agent

import (
	"fmt"
	"github.com/timsims/pamphlet"
)

type chapterContent struct {
	Title   string
	Content string
}

func fetchChapters(path string) ([]chapterContent, error) {
	parser, err := pamphlet.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening pdf at %s: %w", path, err)
	}

	book := parser.GetBook()

	chaptersContent := make([]chapterContent, len(book.Chapters), len(book.Chapters))
	chapters := book.Chapters
	for i, chapter := range chapters {
		chapContent, err := chapter.GetContent()
		if err != nil {
			return nil, fmt.Errorf("failed to get content for chapter %d: %w", i+1, err)
		}
		chaptersContent[i] = chapterContent{
			Title:   chapter.Title,
			Content: chapContent,
		}
	}
	return chaptersContent, nil
}
