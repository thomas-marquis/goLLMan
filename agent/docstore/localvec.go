package docstore

import (
	"context"
	"fmt"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/localvec"
	"github.com/thomas-marquis/goLLMan/pkg"
	"github.com/tmc/langchaingo/textsplitter"
	"strings"
)

type LocalVecStore struct {
	store     *localvec.DocStore
	retriever ai.Retriever
	splitter  textsplitter.TextSplitter
}

func NewLocalVecStore(g *genkit.Genkit, embeddingModel string) (*LocalVecStore, error) {
	splitModel := strings.Split(embeddingModel, "/")
	if len(splitModel) != 2 {
		return nil, fmt.Errorf("embedding model must be in the format 'provider/model', got: %s", embeddingModel)
	}
	provider := splitModel[0]
	model := splitModel[1]

	docStore, retriever, err := localvec.DefineRetriever(g, "docStoreRetriever",
		localvec.Config{Embedder: genkit.LookupEmbedder(g, provider, model)},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create local vector store: %w", err)
	}

	splitter := textsplitter.NewRecursiveCharacter(
		textsplitter.WithChunkSize(1000),
		textsplitter.WithChunkOverlap(20),
	)

	if err := localvec.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize local vector store: %w", err)
	}

	return &LocalVecStore{
		store:     docStore,
		retriever: retriever,
		splitter:  splitter,
	}, nil
}

func (s *LocalVecStore) Index(ctx context.Context, docs []*ai.Document) error {
	preparedDocs := make([]*ai.Document, 0)
	for _, doc := range docs {
		text := pkg.ContentToText(doc.Content)

		chunks, err := s.splitter.SplitText(text)
		if err != nil {
			return err
		}

		for _, chunk := range chunks {
			if len(chunk) == 0 {
				continue
			}
			chunk = strings.TrimSpace(chunk)
			preparedDocs = append(preparedDocs, ai.DocumentFromText(chunk, nil))
		}
	}

	documentsBatches := make([][]*ai.Document, 0)
	batchSize := 10
	for i := 0; i < len(preparedDocs); i += batchSize {
		end := i + batchSize
		if end > len(preparedDocs) {
			end = len(preparedDocs)
		}
		documentsBatches = append(documentsBatches, preparedDocs[i:end])
	}

	for _, batch := range documentsBatches {
		if err := localvec.Index(ctx, batch, s.store); err != nil {
			return fmt.Errorf("failed to index batch of documents: %w", err)
		}
	}
	return nil
}

func (s *LocalVecStore) Retrieve(ctx context.Context, query string, limit int) ([]*ai.Document, error) {
	retrieverResponse, err := ai.Retrieve(ctx, s.retriever, ai.WithTextDocs(query))
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve documents: %w", err)
	}

	foundDocs := retrieverResponse.Documents

	if len(foundDocs) >= limit {
		foundDocs = foundDocs[:limit]
	}

	return foundDocs, nil
}

func getTitle(doc *ai.Document) string {
	var title string
	if value, exists := doc.Metadata["title"]; exists {
		title = value.(string)
	} else {
		title = ""
	}

	return title
}
