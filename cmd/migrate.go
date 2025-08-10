package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/thomas-marquis/goLLMan/internal/infrastructure/migrations"
	"github.com/thomas-marquis/goLLMan/internal/infrastructure/orm"
	"os"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Manage database migrations",
	Long:  "Perform database migrations to set up the database schema",
	Run: func(cmd *cobra.Command, args []string) {
		dsn := os.Getenv("DATABASE_URL")
		if dsn == "" {
			fmt.Println("Error: DATABASE_URL environment variable not set")
			os.Exit(1)
		}

		db, err := orm.NewGormDB(dsn)
		if err != nil {
			fmt.Printf("Error connecting to database: %v\n", err)
			os.Exit(1)
		}

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
