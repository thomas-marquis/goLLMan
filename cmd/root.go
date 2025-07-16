package cmd

import (
	"context"
	"fmt"
	"github.com/firebase/genkit/go/genkit"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pgxvec "github.com/pgvector/pgvector-go/pgx"
	"github.com/thomas-marquis/genkit-mistral/mistral"
	"github.com/thomas-marquis/goLLMan/agent"
	"github.com/thomas-marquis/goLLMan/agent/docstore"
	"github.com/thomas-marquis/goLLMan/agent/loader"
	"github.com/thomas-marquis/goLLMan/agent/session"
	"github.com/thomas-marquis/goLLMan/agent/session/in_memory"
	"github.com/thomas-marquis/goLLMan/internal/domain"
	"github.com/thomas-marquis/goLLMan/internal/infrastructure"
	"github.com/thomas-marquis/goLLMan/pkg"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	embeddingModel  = "mistral/mistral-embed"
	completionModel = "mistral/mistral-small"
)

var (
	cfgFile     string
	agentConfig agent.Config

	bookRepository domain.BookRepository
	vectorStore    docstore.DocStore
	mainAgent      *agent.Agent
	sessionStore   session.Store

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

	rootCmd.PersistentFlags().StringP("docstore", "s", "local",
		"document store implementation to use (local, pgvector)")
	viper.BindPFlag("agent.docstore", rootCmd.PersistentFlags().Lookup("docstore"))

	rootCmd.PersistentFlags().BoolP("disable-ai", "d", false,
		"Disable AI and use generic fake response instead (for testing purpose).")
	viper.BindPFlag("agent.disableAI", rootCmd.PersistentFlags().Lookup("disable-ai"))

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

	switch viper.GetString("agent.docstore") {
	case docstore.DocStoreTypePgvector:
		p := pgConfig{
			DbName:   viper.GetString("postgres.database"),
			User:     viper.GetString("postgres.user"),
			Password: viper.GetString("postgres.password"),
			Host:     viper.GetString("postgres.host"),
			Port:     viper.GetString("postgres.port"),
		}

		pool, err := initPgPool(p)
		if err != nil {
			rootCmd.Printf("Error initializing PostgreSQL pool: %s\n", err)
			os.Exit(1)
		}

		bookRepository = infrastructure.NewBookRepositoryPostgres(pool)

		vectorStore, err = docstore.NewPgVectorStore(g, pool, bookRepository, embeddingModel)
		if err != nil {
			rootCmd.Printf("Error creating pgvector store: %s\n", err)
			return
		}

	case docstore.DocStoreTypeLocal:
		bookRepository = infrastructure.NewBookRepositoryLocal(viper.GetString("local.booksJsonPath"))
		vectorStore, err = docstore.NewLocalVecStore(g, bookRepository, embeddingModel)
		if err != nil {
			rootCmd.Printf("Error creating local vector store: %s\n", err)
			return
		}

	default:
		rootCmd.Printf("Unknown document store implementation: %s\n", viper.GetString("agent.docstore"))
		return
	}

	agentConfig = agent.Config{
		SessionID:           viper.GetString("session"),
		Verbose:             viper.GetBool("verbose"),
		DockStoreImpl:       viper.GetString("agent.docstore"),
		DisableAI:           viper.GetBool("agent.disableAI"),
		SessionMessageLimit: viper.GetInt("agent.sessionMessageLimit"),
	}

	docLoader := loader.NewLocalEpubLoader(bookRepository)

	sessionStore = in_memory.NewSessionStore()

	mainAgent = agent.New(g, agentConfig, sessionStore, docLoader, bookRepository, vectorStore)

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

func initPgPool(cfg pgConfig) (*pgxpool.Pool, error) {
	pgUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DbName)

	pgConfig, err := pgxpool.ParseConfig(pgUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PostgreSQL URL: %w", err)
	}

	pgConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		return pgxvec.RegisterTypes(ctx, conn)
	}

	ctx := context.Background()
	pool, err := pgxpool.NewWithConfig(ctx, pgConfig)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}
	pkg.Logger.Println("Connected to PostgreSQL")

	return pool, nil
}
