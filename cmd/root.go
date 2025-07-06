package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	rootCmd = &cobra.Command{
		Use:   "goLLMan",
		Short: "A golang implementation of an agentic intelligent tinking program.",
		Long: `This applicaiton is able to thing by itself and make decisions.

    It's an experimental project that aims to create an agentic intelligent thinking program using Go.
    A possible side effect could be the AI-world domination, so use it with caution.`,
	}
)

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "secret.yaml",
		"config file (default is ./secret.yaml)")

	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "enable verbose output")
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))

	rootCmd.AddCommand(chatCmd)
	rootCmd.AddCommand(indexCmd)
}

func initConfig() {
	viper.SetConfigFile(cfgFile)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		rootCmd.Printf("Error reading config file: %s\n", viper.ConfigFileUsed())
		return
	}
}
