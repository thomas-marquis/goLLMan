package agent

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	genkit_core "github.com/firebase/genkit/go/core"
)

type cmdLineController struct {
	flow *genkit_core.Flow[string, string, struct{}]
}

func NewCmdLineController(flow *genkit_core.Flow[string, string, struct{}]) *cmdLineController {
	return &cmdLineController{flow}
}

func (c *cmdLineController) Run() error {
	var ctx context.Context

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Enter a command (or 'exit' or 'quit' to quit):")
	for {
		ctx = context.Background()
		fmt.Println("\n## User:")
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("Failed to read input: %v", err)
		}

		input = strings.TrimSuffix(input, "\n")
		if input == "exit" || input == "quit" {
			fmt.Println("## AI/\nSee you next time!")
			break
		}

		fmt.Println("## AI:")
		result, err := c.flow.Run(ctx, input)
		if err != nil {
			return fmt.Errorf("Failed to generate response from flow: %v", err)
		}
		fmt.Println(strings.TrimSuffix(result, "\n"))
	}
	return nil
}
