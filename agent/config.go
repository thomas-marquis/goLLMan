package agent

import "time"

type Config struct {
	SessionID           string
	Verbose             bool
	SessionMessageLimit int
	DisableAI           bool

	EmbeddingVectorSize int

	DockStoreImpl string

	MistralApiKey               string
	MistralMaxRequestsPerSecond int
	MistralTimeout              time.Duration
}
