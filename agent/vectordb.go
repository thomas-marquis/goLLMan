package agent

import (
	"context"
	"fmt"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
)

const (
	retrieveQuery = `SELECT id, content
FROM book_index
WHERE book_id = $1
ORDER BY embedding <#> $2
LIMIT $3`
)

type dbResult struct {
}

func definePgVectorRetriever(ctx context.Context, g *genkit.Genkit, cfg Config, embedder ai.Embedder) (ai.Retriever, error) {
	pgUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		cfg.PgUser, cfg.PgPassword, cfg.PgHost, cfg.PgPort, cfg.PgDatabaseName)

	pool, err := pgxpool.New(ctx, pgUrl)
	if err != nil {
		return nil, err
	}

	return genkit.DefineRetriever(g, "pgvector", "pgVectorRetriever",
		func(ctx context.Context, req *ai.RetrieverRequest) (*ai.RetrieverResponse, error) {
			eres, err := ai.Embed(ctx, embedder, ai.WithDocs(req.Query))
			if err != nil {
				return nil, err
			}

			vec := pgvector.NewVector(eres.Embeddings[0].Embedding)

			bookId, ok := req.Query.Metadata["bookId"].(string)
			if !ok {
				return nil, fmt.Errorf("bookId not found in query metadata")
			}

			nbDocs, ok := req.Query.Metadata["nbDocs"].(int)
			if !ok {
				return nil, fmt.Errorf("nbDocs not found in query metadata")
			}

			row := pool.QueryRow(ctx, retrieveQuery, bookId, vec, nbDocs)

		}), nil
}
