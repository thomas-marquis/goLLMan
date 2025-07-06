package cmd

import (
	"github.com/spf13/viper"
	"github.com/thomas-marquis/goLLMan/agent"

	"github.com/spf13/cobra"
)

// chatCmd represents the chat command
var (
	controllerType string
	chatCmd        = &cobra.Command{
		Use:   "chat",
		Short: "Chat in the terminal.",
		Long:  `Interactive terminal chat interface. Type "quit" to exit.`,
		Run: func(cmd *cobra.Command, args []string) {
			apiToken := viper.GetString("mistral.apiToken")
			ctrlTypeValue, _ := cmd.Flags().GetString("interface")
			ctrlType, err := agent.CtrlTypeFromString(ctrlTypeValue)
			if err != nil {
				cmd.Println("Error getting interface type:", err)
				return
			}

			a := agent.New()
			if err := a.Bootstrap(apiToken, ctrlType); err != nil {
				cmd.Println("Error bootstrapping agent:", err)
				return
			}

			if err := a.StartChatSession(); err != nil {
				cmd.Println("Error running chat session:", err)
				return
			}
		},
	}
)

func init() {
	chatCmd.Flags().StringVarP(&controllerType, "interface", "i", "cmd",
		"Interface type to use for the chat session. Options: cmd, http.")
}
