package loader

import "github.com/firebase/genkit/go/ai"

type DocumentLoader interface {
	Load() ([]*ai.Document, error)
}
