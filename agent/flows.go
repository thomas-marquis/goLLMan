package agent

import (
	"context"
	"fmt"
	"github.com/thomas-marquis/genkit-mistral/mistral"
	"log"
	"os"
	"time"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

var (
	logger = log.New(os.Stdout, "goLLMan: ", log.LstdFlags|log.Lshortfile)
)

func Bootstrap(apiToken string, controllerType ControllerType) (Controller, error) {
	ctx := context.Background()
	g, err := genkit.Init(ctx,
		genkit.WithPlugins(
			mistral.NewPlugin(apiToken,
				mistral.WithRateLimiter(mistral.NewBucketCallsRateLimiter(1, 1, time.Second))),
		),
		genkit.WithDefaultModel("mistral/mistral-small"),
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize Genkit: %w", err)
	}

	chatFlow := genkit.DefineFlow(g, "chatFlow",
		func(ctx context.Context, input string) (string, error) {
			resp, err := genkit.Generate(ctx, g,
				ai.WithSystem("You are a silly assistant."),
				ai.WithPrompt(input),
			)
			if err != nil {
				return "", fmt.Errorf("failed to generate response: %w", err)
			}
			return resp.Text(), nil
		})

	switch controllerType {
	case CtrlTypeCmdLine:
		return NewCmdLineController(chatFlow), nil
	case CtrlTypeHTTP:
		return NewHTTPController(chatFlow), nil
	}

	panic("invalid controller type: " + string(controllerType))
}
