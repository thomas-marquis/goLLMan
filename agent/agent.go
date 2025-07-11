package agent

import (
	"context"
	"fmt"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
	"github.com/thomas-marquis/genkit-mistral/mistral"
	"github.com/thomas-marquis/goLLMan/agent/docstore"
	"github.com/thomas-marquis/goLLMan/agent/loader"
	"github.com/thomas-marquis/goLLMan/agent/session"
	"github.com/thomas-marquis/goLLMan/pkg"
	"time"
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
	docLoader       loader.DocumentLoader
	docStore        docstore.DocStore
	store           session.Store
	cfg             Config
}

func New(cfg Config, store session.Store) *Agent {
	return &Agent{
		store: store,
		cfg:   cfg,
	}
}

func (a *Agent) Flow() *core.Flow[ChatbotInput, string, struct{}] {
	if a.chatbotAIFlow == nil {
		panic("Flow called on a nil flow: please call Bootstrap first")
	}
	if a.cfg.DisableAI {
		return a.chatbotFakeFlow
	}
	return a.chatbotAIFlow
}

func (a *Agent) Genkit() *genkit.Genkit {
	if a.g == nil {
		panic("Genkit called on a nil Genkit: please call Bootstrap first")
	}
	return a.g
}

func (a *Agent) Bootstrap(apiToken string) error {
	ctx := context.Background()
	var err error
	a.g, err = genkit.Init(ctx,
		genkit.WithPlugins(
			mistral.NewPlugin(apiToken,
				mistral.WithRateLimiter(mistral.NewBucketCallsRateLimiter(6, 6, time.Second)),
				mistral.WithVerbose(a.cfg.Verbose),
			),
		),
		genkit.WithDefaultModel("mistral/mistral-small"),
	)
	if err != nil {
		return fmt.Errorf("failed to initialize Genkit: %w", err)
	}

	a.docLoader = loader.NewLocalEpubLoader("documents/effectiveconcurrencyingo.epub")
	a.docStore, err = docstore.NewLocalVecStore(a.g, "mistral/mistral-embed")
	if err != nil {
		return fmt.Errorf("failed to create local vector docStore: %w", err)
	}

	a.indexerFlow = genkit.DefineFlow(a.g, "indexerFlow", a.indexerFlowHandler)
	a.chatbotAIFlow = genkit.DefineFlow(a.g, "chatbotAIFlow", a.chatbotAiFlowHandler)
	a.chatbotFakeFlow = genkit.DefineFlow(a.g, "chatbotFakeFlow", a.chatbotFakeHandle)

	return nil
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
