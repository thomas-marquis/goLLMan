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
	Session  string `json:"session,omitempty"`
}

type Agent struct {
	g               *genkit.Genkit
	indexerFlow     *core.Flow[string, any, struct{}]
	chatbotAIFlow   *core.Flow[ChatbotInput, string, struct{}]
	chatbotFakeFlow *core.Flow[ChatbotInput, string, struct{}]
	docLoader       loader.BookLoader
	docStore        docstore.DocStore
	store           session.Store
	cfg             Config
	bookRepository  domain.BookRepository
}

func New(
	g *genkit.Genkit,
	cfg Config,
	store session.Store,
	docLoader loader.BookLoader,
	bookRepository domain.BookRepository,
) *Agent {
	a := &Agent{
		g:              g,
		store:          store,
		cfg:            cfg,
		bookRepository: bookRepository,
		docLoader:      docLoader,
	}

	a.indexerFlow = genkit.DefineFlow(a.g, "indexerFlow", a.indexerFlowHandler)
	a.chatbotAIFlow = genkit.DefineFlow(a.g, "chatbotAIFlow", a.chatbotAiFlowHandler)
	a.chatbotFakeFlow = genkit.DefineFlow(a.g, "chatbotFakeFlow", a.chatbotFakeHandle)

	return a
}

func (a *Agent) Flow() *core.Flow[ChatbotInput, string, struct{}] {
	if a.cfg.DisableAI {
		return a.chatbotFakeFlow
	}
	if a.chatbotAIFlow == nil {
		panic("Flow called on a nil flow: please call Bootstrap first")
	}
	return a.chatbotAIFlow
}

func (a *Agent) G() *genkit.Genkit {
	if a.g == nil {
		panic("G called on a nil G: please call Bootstrap first")
	}
	return a.g
}

func (a *Agent) Index() error {
	if a.indexerFlow == nil {
		panic("Index called on a nil flow: please call Bootstrap first")
	}

	ctx := context.Background()
	pkg.Logger.Println("Indexing started...")
	_, err := a.indexerFlow.Run(ctx, "documents/effectiveconcurrencyingo.epub")
	if err != nil {
		return fmt.Errorf("failed to run indexer flow: %w", err)
	}
	pkg.Logger.Println("Indexing flow complete")
	return nil
}
