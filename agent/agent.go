package agent

import (
	"context"
	"log"
	"os"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/thomas-marquis/goLLMan/mistral"
)

var (
	logger = log.New(os.Stdout, "goLLMan: ", log.LstdFlags)
)

func Run(apiToken string) {
	ctx := context.Background()
	g, err := genkit.Init(ctx,
		genkit.WithPlugins(
			mistral.NewPlugin(apiToken),
		),
	)
	if err != nil {
		logger.Fatalf("Failed to initialize Genkit: %v", err)
	}

	resp, err := genkit.Generate(ctx, g,
		ai.WithModelName("mistral/mistral-large"),
		ai.WithPrompt("What is the capital of France?"),
	)
	if err != nil {
		logger.Fatalf("Failed to generate response: %v", err)
	}
	logger.Printf("Response: %s", resp.Text())
}
