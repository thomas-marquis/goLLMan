package agent

import (
	"context"
	"fmt"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
	"github.com/thomas-marquis/genkit-mistral/mistral"
	"github.com/thomas-marquis/goLLMan/agent/loader"
	"github.com/thomas-marquis/goLLMan/agent/session"
	"github.com/thomas-marquis/goLLMan/internal/domain"
	"github.com/thomas-marquis/goLLMan/pkg"
	"time"
)

type ChatbotInput struct {
	Question string `json:"question"`
	Session  string `json:"session,omitempty"`
}

type Agent struct {
	g               *genkit.Genkit
	indexerFlow     *core.Flow[domain.Book, any, struct{}]
	chatbotFlow     *core.Flow[ChatbotInput, string, struct{}]
	docLoader       loader.BookLoader
	bookVectorStore domain.BookVectorStore
	sessionStore    session.Store
	cfg             Config
	bookRepository  domain.BookRepository
	retriever       ai.Retriever
	embedder        ai.Embedder
}

func New(
	cfg Config,
	store session.Store,
	docLoader loader.BookLoader,
	bookRepository domain.BookRepository,
	bookVectorStore domain.BookVectorStore,
) *Agent {

	ctx := context.Background()

	rateLimit := cfg.MistralMaxRequestsPerSecond
	g, err := genkit.Init(ctx,
		genkit.WithPlugins(
			mistral.NewPlugin(cfg.MistralApiKey,
				mistral.WithRateLimiter(mistral.NewBucketCallsRateLimiter(rateLimit, rateLimit, time.Second)),
				mistral.WithVerbose(cfg.Verbose),
				mistral.WithClientTimeout(cfg.MistralTimeout),
			),
		),
		genkit.WithDefaultModel(cfg.CompletionModel),
	)
	if err != nil {
		pkg.Logger.Fatalf("failed to init genkit: %s", err)
	}

	a := &Agent{
		g:               g,
		sessionStore:    store,
		cfg:             cfg,
		bookRepository:  bookRepository,
		docLoader:       docLoader,
		bookVectorStore: bookVectorStore,
	}

	a.indexerFlow = genkit.DefineFlow(g, "indexerFlow", a.indexerFlowHandler)
	a.retriever = genkit.DefineRetriever(g, "gollman", "bookRetriever", a.bookRetrieverHandler)

	embProvider, embModel, err := pkg.ParseModelRef(cfg.EmbeddingModel)
	if err != nil {
		pkg.Logger.Fatalf("failed to parse embedding model: %s", err)
	}
	a.embedder = genkit.LookupEmbedder(g, embProvider, embModel)

	a.chatbotFlow = genkit.DefineFlow(g, "chatbotAIFlow", a.chatbotAiFlowHandler)

	return a
}

// Flow returns the chatbot flow used by the agent.
func (a *Agent) Flow() *core.Flow[ChatbotInput, string, struct{}] {
	return a.chatbotFlow
}

// G returns the Genkit instance used by the agent.
func (a *Agent) G() *genkit.Genkit {
	return a.g
}

// Index starts the indexing process for the agent.
func (a *Agent) Index() error {
	ctx := context.Background()
	//filePath := "documents/effectiveconcurrencyingo.epub"
	filePath := "documents/votreidee.epub"
	parsedBook, err := a.bookRepository.ReadFromFile(ctx, filePath)
	if err != nil {
		return fmt.Errorf("failed to read parsedBook from file %s: %w", filePath, err)
	}

	book, err := a.bookRepository.Add(ctx, parsedBook.Title, parsedBook.Author, parsedBook.Metadata)
	if err != nil {
		return fmt.Errorf("failed to add book %s by %s: %w", parsedBook.Title, parsedBook.Author, err)
	}

	pkg.Logger.Println("Indexing started...")
	if _, err := a.indexerFlow.Run(ctx, book); err != nil {
		return fmt.Errorf("failed to run indexer flow: %w", err)
	}

	pkg.Logger.Println("Indexing flow complete")
	return nil
}
