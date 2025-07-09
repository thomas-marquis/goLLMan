package cmd

import (
	"github.com/spf13/viper"
	"github.com/thomas-marquis/goLLMan/agent"
	"github.com/thomas-marquis/goLLMan/agent/session/in_memory"
	"github.com/thomas-marquis/goLLMan/controller"
	"github.com/thomas-marquis/goLLMan/controller/cmdline"
	"github.com/thomas-marquis/goLLMan/controller/server"
	"os"

	"github.com/spf13/cobra"
)

// chatCmd represents the chat command
var (
	controllerType string
	sessionID      string
	chatCmd        = &cobra.Command{
		Use:   "chat",
		Short: "Chat in the terminal.",
		Long:  `Interactive terminal chat interface. Type "quit" to exit.`,
		Run: func(cmd *cobra.Command, args []string) {
			apiToken := viper.GetString("mistral.apiToken")
			verbose := viper.GetBool("verbose")
			ctrlTypeValue, _ := cmd.Flags().GetString("interface")
			ctrlType, err := controller.CtrlTypeFromString(ctrlTypeValue)
			if err != nil {
				cmd.Println("Error getting interface type:", err)
				return
			}

			store := in_memory.NewSessionStore()

			agentCfg := agent.Config{
				SessionID:           viper.GetString("session"),
				Verbose:             verbose,
				SessionMessageLimit: 6,
			}

			a := agent.New(agentCfg, store)
			if err := a.Bootstrap(apiToken); err != nil {
				cmd.Println("Error bootstrapping agent:", err)
				os.Exit(1)
			}

			var ctrl controller.Controller
			switch ctrlType {
			case controller.CtrlTypeCmdLine:
				ctrl = cmdline.New(agentCfg, a.Flow())
			case controller.CtrlTypeHTTP:
				ctrl = server.New(agentCfg, a.Flow(), store, a.Genkit())
			default:
				cmd.Println("unsupported controller type: %s", controllerType)
				os.Exit(1)
			}

			if err := ctrl.Run(); err != nil {
				cmd.Println("Error running chat session:", err)
				os.Exit(1)
			}
		},
	}
)

func init() {
	chatCmd.Flags().StringVarP(&controllerType, "interface", "i", "cmd",
		"Interface type to use for the chat session. Options: cmd, http.")

	chatCmd.Flags().StringVarP(&sessionID, "session", "s", "",
		"Session ID to use for the chat session. If not provided, a new session will be created.")
}
