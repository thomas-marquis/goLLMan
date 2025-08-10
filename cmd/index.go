package cmd

import (
	"github.com/spf13/cobra"
)

// indexCmd represents the index command
var (
	indexCmd = &cobra.Command{
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
			if err := mainAgent.Index(); err != nil {
				cmd.Println("an error occurred during indexing:", err)
				return
			}
		},
	}
)

func init() {}
