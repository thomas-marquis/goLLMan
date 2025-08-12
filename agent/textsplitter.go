package agent

import (
	"regexp"
	"strings"

	"github.com/firebase/genkit/go/ai"
	"github.com/thomas-marquis/goLLMan/pkg"
	"github.com/tmc/langchaingo/textsplitter"
)

var (
	atxHeadingRe      = regexp.MustCompile(`^\s{0,3}#{1,6}(\s+|$).*`)
	setextUnderlineRe = regexp.MustCompile(`^\s{0,3}(=+|-+)\s*$`)
)

func makeTextSplitter() textsplitter.TextSplitter {
	splitter := textsplitter.NewMarkdownTextSplitter(
		textsplitter.WithCodeBlocks(true),
		textsplitter.WithKeepSeparator(true),
		textsplitter.WithChunkSize(1000),
		textsplitter.WithHeadingHierarchy(true),
	)
	return splitter
}

// isATXHeading returns true if the given line is an ATX-style heading (#, ##, ..., ######).
func isATXHeading(line string) bool {
	return atxHeadingRe.MatchString(line)
}

// isSetextHeading returns true if the given line is a Setext-style heading (=== or ---).
func isSetextHeading(currLineIdx *int, lines []string) bool {
	j := *currLineIdx + 1
	for j < len(lines) && strings.TrimSpace(lines[j]) == "" {
		j++
	}
	if j < len(lines) && setextUnderlineRe.MatchString(strings.TrimSpace(lines[j])) {
		// Treat as Setext heading; skip both lines (and any blanks in between already handled)
		*currLineIdx = j + 1
		return true
	}

	return false
}

// OnlyContainsHeaders returns true if the given text contains only headers (atx or setext).
func OnlyContainsHeaders(text string) bool {
	s := strings.TrimSpace(text)
	if s == "" {
		// Whitespace-only is considered as containing no non-heading content.
		return true
	}

	lines := strings.Split(s, "\n")
	i := 0
	for i < len(lines) {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			i++
			continue
		}

		if isATXHeading(line) {
			i++
			continue
		}
		if isSetextHeading(&i, lines) {
			continue
		}
		return false
	}
	return true
}

// SplitDocuments splits the given documents into chunks based on the given splitter.
func SplitDocuments(splitter textsplitter.TextSplitter, docs []*ai.Document) ([]*ai.Document, error) {
	preparedDocs := make([]*ai.Document, 0)
	for _, doc := range docs {
		text := pkg.ContentToText(doc.Content)
		chunks, err := splitter.SplitText(text)
		if err != nil {
			return nil, err
		}

		for _, chunk := range chunks {
			if len(chunk) == 0 || OnlyContainsHeaders(chunk) {
				continue
			}
			chunk = strings.TrimSpace(chunk)
			preparedDocs = append(preparedDocs, ai.DocumentFromText(chunk, doc.Metadata))
		}
	}
	return preparedDocs, nil
}

func batchDocuments(docs []*ai.Document, batchSize int) [][]*ai.Document {
	if batchSize <= 0 {
		batchSize = 1
	}

	batches := make([][]*ai.Document, 0)
	for i := 0; i < len(docs); i += batchSize {
		end := i + batchSize
		if end > len(docs) {
			end = len(docs)
		}
		batches = append(batches, docs[i:end])
	}
	return batches
}
