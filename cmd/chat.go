package cmd

import (
	"github.com/spf13/viper"
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

	chatCmd = &cobra.Command{
		Use:   "chat",
		Short: "Chat in the terminal.",
		Long:  `Interactive terminal chat interface. Type "quit" to exit.`,
		Run: func(cmd *cobra.Command, args []string) {
			ctrlTypeValue, _ := cmd.Flags().GetString("interface")
			ctrlType, err := controller.CtrlTypeFromString(ctrlTypeValue)
			if err != nil {
				cmd.Println("Error getting interface type:", err)
				return
			}

			agentConfig.SessionID = viper.GetString("session")
			agentConfig.SessionMessageLimit = 6

			var ctrl controller.Controller
			switch ctrlType {
			case controller.CtrlTypeCmdLine:
				ctrl = cmdline.New(agentConfig, mainAgent.Flow())
			case controller.CtrlTypeHTTP:
				ctrl = server.New(agentConfig, mainAgent.Flow(), sessionStore, mainAgent.G())
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

	chatCmd.Flags().StringVarP(&sessionID, "session", "S", "",
		"Session ID to use for the chat session. If not provided, a new session will be created.")
}
