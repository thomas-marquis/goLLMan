package docstore

import (
	"github.com/firebase/genkit/go/ai"
	"github.com/thomas-marquis/goLLMan/pkg"
	"github.com/tmc/langchaingo/textsplitter"
	"strings"
)

func makeTextSplitter() textsplitter.TextSplitter {
	splitter := textsplitter.NewRecursiveCharacter(
		textsplitter.WithChunkSize(1000),
		textsplitter.WithChunkOverlap(20),
	)
	return splitter
}

func splitDocuments(splitter textsplitter.TextSplitter, docs []*ai.Document) ([]*ai.Document, error) {
	preparedDocs := make([]*ai.Document, 0)
	for _, doc := range docs {
		text := pkg.ContentToText(doc.Content)
		chunks, err := splitter.SplitText(text)
		if err != nil {
			return nil, err
		}

		for _, chunk := range chunks {
			if len(chunk) == 0 {
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
