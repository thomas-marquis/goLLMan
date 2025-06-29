package agent

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/thomas-marquis/goLLMan/mistral"
)

var (
	logger = log.New(os.Stdout, "goLLMan: ", log.LstdFlags|log.Lshortfile)
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

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Enter a command (or 'exit' or 'quit' to quit):")
	for {
		fmt.Println("\n## User:")
		input, err := reader.ReadString('\n')
		if err != nil {
			logger.Fatalf("Failed to read input: %v", err)
		}

		input = strings.TrimSuffix(input, "\n")
		if input == "exit" || input == "quit" {
			fmt.Println("## AI/\nSee you next time!")
			break
		}

		fmt.Println("## AI:")
		resp, err := genkit.Generate(ctx, g,
			ai.WithModelName("mistral/mistral-large"),
			// ai.WithSystem("You are a silly assistant."),
			ai.WithPrompt(input),
		)
		if err != nil {
			logger.Fatalf("Failed to generate response: %v", err)
		}
		fmt.Println(strings.TrimSuffix(resp.Text(), "\n"))
	}
}
