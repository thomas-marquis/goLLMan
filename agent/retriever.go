package agent

import (
	"context"
	"errors"
	"fmt"
	"github.com/firebase/genkit/go/ai"
	"github.com/thomas-marquis/goLLMan/internal/domain"
	"github.com/thomas-marquis/goLLMan/pkg"
)

func (a *Agent) bookRetrieverHandler(ctx context.Context, req *ai.RetrieverRequest) (*ai.RetrieverResponse, error) {
	books, err := a.bookRepository.ListSelected(ctx)
	if err != nil {
		if errors.Is(err, domain.ErrBookNotFound) {
			pkg.Logger.Println("No book selected, returning empty retriever response")
			return &ai.RetrieverResponse{Documents: make([]*ai.Document, 0)}, nil
		}
		return nil, fmt.Errorf("failed to get selected books: %w", err)
	}

	limit, ok := req.Query.Metadata["limit"].(int)
	if !ok || limit <= 0 {
		limit = 3
		pkg.Logger.Printf("limit not found in query metadata, applying default limit: %d", limit)
	}

	query := pkg.ContentToText(req.Query.Content)

	docs, err := retrieveDocument(a.embedder, a.bookVectorStore, ctx, books, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve documents: %w", err)
	}

	return &ai.RetrieverResponse{Documents: docs}, nil
}

func retrieveDocument(
	embedder ai.Embedder,
	vectorStore domain.BookVectorStore,
	ctx context.Context,
	books []domain.Book, query string, limit int,
) ([]*ai.Document, error) {
	eReq := &ai.EmbedRequest{
		Input: []*ai.Document{ai.DocumentFromText(query, nil)},
	}
	eRes, err := embedder.Embed(ctx, eReq)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}
	vec := eRes.Embeddings[0].Embedding

	return vectorStore.Retrieve(ctx, books, vec, limit)
}
