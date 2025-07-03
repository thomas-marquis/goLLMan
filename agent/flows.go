package agent

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/thomas-marquis/goLLMan/mistral"
)

var (
	logger = log.New(os.Stdout, "goLLMan: ", log.LstdFlags|log.Lshortfile)
)

func Bootstrap(apiToken string, ctrlImpltType ControllerType) (Controller, error) {
	ctx := context.Background()
	g, err := genkit.Init(ctx,
		genkit.WithPlugins(
			mistral.NewPlugin(apiToken),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize Genkit: %w", err)
	}

	chatFlow := genkit.DefineFlow(g, "chatFlow",
		func(ctx context.Context, input string) (string, error) {
			resp, err := genkit.Generate(ctx, g,
				ai.WithModelName("mistral/mistral-large"),
				// ai.WithSystem("You are a silly assistant."),
				ai.WithPrompt(input),
			)
			if err != nil {
				return "", fmt.Errorf("failed to generate response: %w", err)
			}
			return resp.Text(), nil
		})

	switch ctrlImpltType {
	case CtrlTypeCmdLine:
		return NewCmdLineController(chatFlow), nil
	}

	panic("invalid controller type: " + string(ctrlImpltType))
}
