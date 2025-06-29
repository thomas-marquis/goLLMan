package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/thomas-marquis/goLLMan/agent"
)

var rootCmd = &cobra.Command{
	Use:   "goLLMan",
	Short: "A golang implementation of an agentic intelligent tinking program.",
	Long: `This applicaiton is able to thing by itself and make decisions.

    It's an experimental project that aims to create an agentic intelligent thinking program using Go.
    A possible side effect could be the AI-world domination, so use it with caution.`,
	Run: func(cmd *cobra.Command, args []string) {
		agent.Run()
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
