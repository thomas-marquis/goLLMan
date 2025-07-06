package agent

import (
	"bufio"
	"context"
	"fmt"
	"github.com/thomas-marquis/goLLMan/agent/session"
	"os"
	"strings"

	genkit_core "github.com/firebase/genkit/go/core"
)

type cmdLineController struct {
	flow *genkit_core.Flow[ChatbotInput, string, struct{}]
	cfg  Config
}

func NewCmdLineController(cfg Config, flow *genkit_core.Flow[ChatbotInput, string, struct{}]) *cmdLineController {
	return &cmdLineController{flow, cfg}
}

func (c *cmdLineController) Run() error {
	var ctx context.Context

	reader := bufio.NewReader(os.Stdin)

	sessionID := c.cfg.SessionID
	if sessionID == "" {
		sessionID = session.GenerateID()
	}

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
		result, err := c.flow.Run(ctx, ChatbotInput{Question: input, Session: sessionID})
		if err != nil {
			return fmt.Errorf("Failed to generate response from flow: %v", err)
		}
		fmt.Println(strings.TrimSuffix(result, "\n"))
	}
	return nil
}
