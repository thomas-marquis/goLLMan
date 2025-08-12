package loader

import (
	"fmt"

	"github.com/JohannesKaufmann/dom"
	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/base"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/commonmark"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/table"
	"github.com/firebase/genkit/go/ai"
	"github.com/thomas-marquis/goLLMan/internal/domain"
	"github.com/timsims/pamphlet"
	"golang.org/x/net/html"
)

type LocalEpubLoader struct {
	repository domain.BookRepository
	conv       *converter.Converter
}

var _ BookLoader = (*LocalEpubLoader)(nil)

func NewLocalEpubLoader(repository domain.BookRepository) *LocalEpubLoader {
	conv := converter.NewConverter(
		converter.WithPlugins(
			base.NewBasePlugin(),
			commonmark.NewCommonmarkPlugin(),
			table.NewTablePlugin(),
		),
	)
	conv.Register.PreRenderer(fixLinksPreRender, converter.PriorityEarly)

	return &LocalEpubLoader{
		repository: repository,
		conv:       conv,
	}
}

func (l *LocalEpubLoader) Load(book domain.Book, file *domain.FileWithContent) ([]*ai.Document, error) {
	parser, err := pamphlet.OpenBytes(file.Content)
	if err != nil {
		return nil, fmt.Errorf("error opening epub content: %w", err)
	}

	parsedBook := parser.GetBook()

	documents := make([]*ai.Document, 0, len(parsedBook.Chapters))
	chapters := parsedBook.Chapters
	for i, chapter := range chapters {
		chapContent, err := chapter.GetContent()
		if err != nil {
			return nil, fmt.Errorf("failed to get content for chapter %d: %w", i+1, err)
		}
		markdown, err := l.conv.ConvertString(chapContent)
		if err != nil {
			return nil, fmt.Errorf("failed to convert chapter %d to markdown: %w", i+1, err)
		}
		if markdown == "" {
			continue
		}
		documents = append(documents, ai.DocumentFromText(markdown, map[string]any{
			"title":   chapter.Title,
			"book_id": book.ID,
		}))
	}

	return documents, nil
}

func fixLinksPreRender(ctx converter.Context, doc *html.Node) {
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			href := dom.GetAttributeOr(n, "href", "")
			hasChildren := n.FirstChild != nil

			if (href == "" || href == "#") && hasChildren {
				n.Data = "div"
				n.Attr = nil
			} else if !hasChildren && href != "" && href != "#" {
				textNode := &html.Node{
					Type: html.TextNode,
					Data: href,
				}
				n.AppendChild(textNode)
			} else {
				n.Data = "p"
				n.Attr = nil
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}

	walk(doc)
}
