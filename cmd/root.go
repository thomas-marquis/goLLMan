package cmd

import (
	"context"
	"fmt"
	"github.com/firebase/genkit/go/genkit"
	"github.com/thomas-marquis/genkit-mistral/mistral"
	"github.com/thomas-marquis/goLLMan/agent"
	"github.com/thomas-marquis/goLLMan/agent/loader"
	"github.com/thomas-marquis/goLLMan/agent/session"
	"github.com/thomas-marquis/goLLMan/agent/session/in_memory"
	"github.com/thomas-marquis/goLLMan/internal/domain"
	"github.com/thomas-marquis/goLLMan/internal/infrastructure"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	completionModel = "mistral/mistral-small"
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

	rootCmd.PersistentFlags().BoolP("disable-ai", "d", false,
		"Disable AI and use generic fake response instead (for testing purpose).")
	viper.BindPFlag("agent.disableAI", rootCmd.PersistentFlags().Lookup("disable-ai"))

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

	ctx := context.Background()

	g, err := initGenkit(
		ctx,
		viper.GetString("mistral.apiKey"),
		viper.GetBool("verbose"),
		viper.GetDuration("mistral.timeout"),
		viper.GetInt("mistral.maxReqPerSec"),
	)
	if err != nil {
		rootCmd.Printf("Error initializing G: %s\n", err)
		return
	}

	p := pgConfig{
		DbName:   viper.GetString("postgres.database"),
		User:     viper.GetString("postgres.user"),
		Password: viper.GetString("postgres.password"),
		Host:     viper.GetString("postgres.host"),
		Port:     viper.GetString("postgres.port"),
	}

	db, err = initPgGormDB(p)
	if err != nil {
		rootCmd.Printf("Error initializing PostgreSQL db: %s\n", err)
		os.Exit(1)
	}

	bookRepoImpl := infrastructure.NewBookRepositoryPostgres(db)
	bookVectorStore = bookRepoImpl
	bookRepository = bookRepoImpl

	agentConfig = agent.Config{
		SessionID:           viper.GetString("session"),
		Verbose:             viper.GetBool("verbose"),
		DockStoreImpl:       viper.GetString("agent.docstore"),
		DisableAI:           viper.GetBool("agent.disableAI"),
		SessionMessageLimit: viper.GetInt("agent.sessionMessageLimit"),
	}

	docLoader := loader.NewLocalEpubLoader(bookRepository)

	sessionStore = in_memory.NewSessionStore()

	mainAgent = agent.New(g, agentConfig, sessionStore, docLoader, bookRepository, bookVectorStore)

}

func initGenkit(ctx context.Context, mistralApiKey string, verbose bool, timeout time.Duration, rateLimit int) (*genkit.Genkit, error) {
	return genkit.Init(ctx,
		genkit.WithPlugins(
			mistral.NewPlugin(mistralApiKey,
				mistral.WithRateLimiter(mistral.NewBucketCallsRateLimiter(rateLimit, rateLimit, time.Second)),
				mistral.WithVerbose(verbose),
				mistral.WithClientTimeout(timeout),
			),
		),
		genkit.WithDefaultModel(completionModel),
	)
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
