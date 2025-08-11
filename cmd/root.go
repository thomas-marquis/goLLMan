package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/thomas-marquis/goLLMan/agent"
	"github.com/thomas-marquis/goLLMan/agent/loader"
	"github.com/thomas-marquis/goLLMan/agent/session"
	"github.com/thomas-marquis/goLLMan/agent/session/in_memory"
	"github.com/thomas-marquis/goLLMan/internal/domain"
	"github.com/thomas-marquis/goLLMan/internal/infrastructure"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
)

const (
	defaultCompletionModel = "mistral/mistral-small"
	defaultEmbeddingModel  = "mistral/mistral-embed"
)

var (
	cfgFile     string
	agentConfig agent.Config

	db *gorm.DB

	mainAgent       *agent.Agent
	sessionStore    session.Store
	bookRepository  domain.BookRepository
	bookVectorStore domain.BookVectorStore

	rootCmd = &cobra.Command{
		Use:   "goLLMan",
		Short: "A golang implementation of an agentic intelligent tinking program.",
		Long: `This applicaiton is able to thing by itself and make decisions.

    It's an experimental project that aims to create an agentic intelligent thinking program using Go.
    A possible side effect could be the AI-world domination, so use it with caution.`,
	}
)

type pgConfig struct {
	DbName   string
	Host     string
	Port     string
	User     string
	Password string
}

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

	viper.SetDefault("agent.retrievalLimit", 6)
	viper.SetDefault("agent.completionModel", defaultCompletionModel)
	viper.SetDefault("agent.embeddingModel", defaultEmbeddingModel)

	rootCmd.AddCommand(chatCmd)
	rootCmd.AddCommand(indexCmd)
	rootCmd.AddCommand(genkitCmd)
}

func initConfig() {
	viper.SetConfigFile(cfgFile)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		rootCmd.Printf("Error reading config file: %s\n", viper.ConfigFileUsed())
		return
	}

	p := pgConfig{
		DbName:   viper.GetString("postgres.database"),
		User:     viper.GetString("postgres.user"),
		Password: viper.GetString("postgres.password"),
		Host:     viper.GetString("postgres.host"),
		Port:     viper.GetString("postgres.port"),
	}

	db, err := initPgGormDB(p)
	if err != nil {
		rootCmd.Printf("Error initializing PostgreSQL db: %s\n", err)
		os.Exit(1)
	}

	bookRepoImpl := infrastructure.NewBookRepositoryPostgres(db)
	bookVectorStore = bookRepoImpl
	bookRepository = bookRepoImpl

	agentConfig = agent.Config{
		SessionID:                   viper.GetString("session"),
		Verbose:                     viper.GetBool("verbose"),
		SessionMessageLimit:         viper.GetInt("agent.sessionMessageLimit"),
		RetrievalLimit:              viper.GetInt("agent.retrievalLimit"),
		MistralApiKey:               viper.GetString("mistral.apiKey"),
		MistralTimeout:              viper.GetDuration("mistral.timeout"),
		MistralMaxRequestsPerSecond: viper.GetInt("mistral.maxReqPerSec"),
		CompletionModel:             viper.GetString("agent.completionModel"),
		EmbeddingModel:              viper.GetString("agent.embeddingModel"),
	}

	docLoader := loader.NewLocalEpubLoader(bookRepository)

	sessionStore = in_memory.NewSessionStore()

	mainAgent = agent.New(agentConfig, sessionStore, docLoader, bookRepository, bookVectorStore)
}

func initPgGormDB(cfg pgConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DbName)

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN: dsn,
	}))
	if err != nil {
		return nil, fmt.Errorf("failed to init postgres GORM instance: %w", err)
	}

	return db, err
}
