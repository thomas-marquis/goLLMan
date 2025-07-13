package docstore

import (
	"fmt"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"strings"
)

func getEmbedder(g *genkit.Genkit, modelRef string) (ai.Embedder, error) {
	splitModel := strings.Split(modelRef, "/")
	if len(splitModel) != 2 {
		return nil, fmt.Errorf("embedding model must be in the format 'provider/model', got: %s", modelRef)
	}
	provider := splitModel[0]
	model := splitModel[1]
	return genkit.LookupEmbedder(g, provider, model), nil
}
