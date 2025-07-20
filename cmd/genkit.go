package cmd

import (
	"github.com/spf13/cobra"
)

// genkitCmd represents the genkit command
var genkitCmd = &cobra.Command{
	Use:   "genkit",
	Short: "Start the genkit UI.",
	Long:  `Start the genkit UI.`,
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func init() {
}
