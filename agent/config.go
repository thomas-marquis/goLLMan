package agent

import "time"

type Config struct {
	SessionID           string
	Verbose             bool
	SessionMessageLimit int

	EmbeddingVectorSize int

	RetrievalLimit int

	MistralApiKey               string
	MistralMaxRequestsPerSecond int
	MistralTimeout              time.Duration

	CompletionModel string
	EmbeddingModel  string
}
