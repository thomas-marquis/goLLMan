package cmd

import (
	"github.com/spf13/viper"
	"github.com/thomas-marquis/goLLMan/agent"
	"github.com/thomas-marquis/goLLMan/agent/session/in_memory"

	"github.com/spf13/cobra"
)

// indexCmd represents the index command
var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Index documents",
	Long: `Index command allows you to index documents for later retrieval.

Documents can be located in the local file system or in a remote location.
See the -l flag documentation for supported locations.
The indexing is the process to embed the document content and then store it in a vector database.
This process may take some time and it usually costs a little if you use a cloud embedding model provider.
The indexing doesn't reindex existing documents.
`,
	Run: func(cmd *cobra.Command, args []string) {
		apiToken := viper.GetString("mistral.apiToken")
		verbose := viper.GetBool("verbose")
		store := in_memory.NewSessionStore()

		agentCfg := agent.Config{Verbose: verbose}

		a := agent.New(agentCfg, store)
		if err := a.Bootstrap(apiToken, agent.CtrlTypeCmdLine); err != nil {
			cmd.Println("Error bootstrapping agent:", err)
			return
		}

		if err := a.Index(); err != nil {
			cmd.Println("an error occurred during indexing:", err)
			return
		}
	},
}

func init() {
}
