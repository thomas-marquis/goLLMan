package agent

import (
	"context"
	"fmt"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
	"github.com/thomas-marquis/goLLMan/agent/docstore"
	"github.com/thomas-marquis/goLLMan/agent/loader"
	"github.com/thomas-marquis/goLLMan/agent/session"
	"github.com/thomas-marquis/goLLMan/internal/domain"
	"github.com/thomas-marquis/goLLMan/pkg"
)

type ChatbotInput struct {
	Question string `json:"question"`
	BookID   string `json:"book_id"`
	Session  string `json:"session,omitempty"`
}

type Agent struct {
	g              *genkit.Genkit
	indexerFlow    *core.Flow[domain.Book, any, struct{}]
	chatbotFlow    *core.Flow[ChatbotInput, string, struct{}]
	docLoader      loader.BookLoader
	docStore       docstore.DocStore
	sessionStore   session.Store
	cfg            Config
	bookRepository domain.BookRepository
}

func New(
	g *genkit.Genkit,
	cfg Config,
	store session.Store,
	docLoader loader.BookLoader,
	bookRepository domain.BookRepository,
	vectorStore docstore.DocStore,
) *Agent {
	a := &Agent{
		g:              g,
		sessionStore:   store,
		cfg:            cfg,
		bookRepository: bookRepository,
		docLoader:      docLoader,
		docStore:       vectorStore,
	}

	a.indexerFlow = genkit.DefineFlow(a.g, "indexerFlow", a.indexerFlowHandler)

	if cfg.DisableAI {
		pkg.Logger.Println("AI disabled")
		a.chatbotFlow = genkit.DefineFlow(a.g, "chatbotFakeFlow", a.chatbotFakeHandle)
	} else {
		a.chatbotFlow = genkit.DefineFlow(a.g, "chatbotAIFlow", a.chatbotAiFlowHandler)
	}

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
