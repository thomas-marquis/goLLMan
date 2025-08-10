package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/thomas-marquis/goLLMan/internal/infrastructure/migrations"
	"os"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Manage database migrations",
	Long:  "Perform database migrations to set up the database schema",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Applying migrations...")
		if err := migrations.MigrateUp(db); err != nil {
			fmt.Printf("Error applying migrations: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Migrations applied successfully!")
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}
