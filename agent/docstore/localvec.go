package docstore

import (
	"context"
	"fmt"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/localvec"
	"github.com/thomas-marquis/goLLMan/internal/domain"
	"github.com/thomas-marquis/goLLMan/pkg"
	"github.com/tmc/langchaingo/textsplitter"
)

type LocalVecStore struct {
	store      *localvec.DocStore
	retriever  ai.Retriever
	splitter   textsplitter.TextSplitter
	repository domain.BookRepository
}

var _ DocStore = (*LocalVecStore)(nil)

func NewLocalVecStore(g *genkit.Genkit, repository domain.BookRepository, embeddingModel string) (*LocalVecStore, error) {
	pkg.Logger.Printf("Initializing local vector store with embedding model: %s\n", embeddingModel)

	embedder, err := getEmbedder(g, embeddingModel)
	if err != nil {
		return nil, err
	}

	docStore, retriever, err := localvec.DefineRetriever(g, "docStoreRetriever",
		localvec.Config{Embedder: embedder},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create local vector store: %w", err)
	}

	splitter := makeTextSplitter()

	if err := localvec.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize local vector store: %w", err)
	}

	return &LocalVecStore{
		store:      docStore,
		retriever:  retriever,
		splitter:   splitter,
		repository: repository,
	}, nil
}

func (s *LocalVecStore) Index(ctx context.Context, book domain.Book, docs []*ai.Document) error {
	preparedDocs, err := splitDocuments(s.splitter, docs)
	if err != nil {
		return fmt.Errorf("failed to split documents: %w", err)
	}

	documentsBatches := batchDocuments(preparedDocs, 10)

	for _, batch := range documentsBatches {
		if err := localvec.Index(ctx, batch, s.store); err != nil {
			return fmt.Errorf("failed to index batch of documents: %w", err)
		}
	}
	return nil
}

func (s *LocalVecStore) Retrieve(ctx context.Context, book domain.Book, query string, limit int) ([]*ai.Document, error) {
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
