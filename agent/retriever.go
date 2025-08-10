package agent

import (
	"context"
	"fmt"
	"github.com/firebase/genkit/go/ai"
	"github.com/thomas-marquis/goLLMan/internal/domain"
	"github.com/thomas-marquis/goLLMan/pkg"
)

func (a *Agent) bookRetrieverHandler(ctx context.Context, req *ai.RetrieverRequest) (*ai.RetrieverResponse, error) {
	book_id, ok := req.Query.Metadata["book_id"].(string)
	if !ok {
		return nil, fmt.Errorf("book not found in query metadata")
	}
	limit, ok := req.Query.Metadata["limit"].(int)
	if !ok || limit <= 0 {
		limit = 3
		pkg.Logger.Printf("limit not found in query metadata, applying default limit: %d", limit)
	}
	query := pkg.ContentToText(req.Query.Content)

	book, err := a.bookRepository.GetByID(ctx, book_id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve book by ID %s: %w", book_id, err)
	}

	docs, err := retrieveDocument(a.embedder, a.bookVectorStore, ctx, book, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve documents: %w", err)
	}

	return &ai.RetrieverResponse{Documents: docs}, nil
}

func retrieveDocument(
	embedder ai.Embedder,
	vectorStore domain.BookVectorStore,
	ctx context.Context,
	book domain.Book, query string, limit int,
) ([]*ai.Document, error) {
	eReq := &ai.EmbedRequest{
		Input: []*ai.Document{ai.DocumentFromText(query, nil)},
	}
	eRes, err := embedder.Embed(ctx, eReq)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}
	vec := eRes.Embeddings[0].Embedding

	return vectorStore.Retrieve(ctx, book, vec, limit)
}
