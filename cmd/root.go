package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/thomas-marquis/goLLMan/agent"
)

type Secret struct {
	Mistral struct {
		ApiToken string `yaml:"apiToken"`
	} `yaml:"mistral"`
}

var rootCmd = &cobra.Command{
	Use:   "goLLMan",
	Short: "A golang implementation of an agentic intelligent tinking program.",
	Long: `This applicaiton is able to thing by itself and make decisions.

    It's an experimental project that aims to create an agentic intelligent thinking program using Go.
    A possible side effect could be the AI-world domination, so use it with caution.`,
	Run: func(cmd *cobra.Command, args []string) {
		viper.SetConfigFile("secret.yaml")
		if err := viper.ReadInConfig(); err != nil {
			cmd.Println("Error reading config file:", err)
			return
		}

		var secrets Secret
		if err := viper.Unmarshal(&secrets); err != nil {
			cmd.Println("Error unmarshalling config:", err)
			return
		}

		ctrl, err := agent.Bootstrap(secrets.Mistral.ApiToken, agent.CtrlTypeCmdLine)
		if err != nil {
			cmd.Println("Error bootstrapping controller:", err)
			return
		}
		if err := ctrl.Run(); err != nil {
			cmd.Println("Error running controller:", err)
			return
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
