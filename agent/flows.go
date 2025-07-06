package agent

import (
	"context"
	"fmt"
	"github.com/firebase/genkit/go/core"
	"github.com/thomas-marquis/genkit-mistral/mistral"
	"github.com/thomas-marquis/goLLMan/agent/docstore"
	"github.com/thomas-marquis/goLLMan/agent/loader"
	"github.com/thomas-marquis/goLLMan/pkg"
	"time"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

type Agent struct {
	g           *genkit.Genkit
	indexerFlow *core.Flow[string, any, struct{}]
	chatbotFlow *core.Flow[string, string, struct{}]
	docLoader   loader.DocumentLoader
	docStore    docstore.DocStore
	ctrl        Controller
	verbose     bool
}

func New(verbose bool) *Agent {
	return &Agent{
		verbose: verbose,
	}
}

func (a *Agent) Bootstrap(apiToken string, controllerType ControllerType) error {
	ctx := context.Background()
	var err error
	a.g, err = genkit.Init(ctx,
		genkit.WithPlugins(
			mistral.NewPlugin(apiToken,
				mistral.WithRateLimiter(mistral.NewBucketCallsRateLimiter(6, 6, time.Second)),
				mistral.WithVerbose(a.verbose),
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

	a.indexerFlow = genkit.DefineFlow(
		a.g, "indexerFlow",
		func(ctx context.Context, path string) (any, error) {
			documents, err := genkit.Run(ctx, "loadDocuments", func() ([]*ai.Document, error) {
				return a.docLoader.Load()
			})
			if err != nil {
				return nil, fmt.Errorf("failed to load documents from path %s: %w", path, err)
			}

			_, err = genkit.Run(ctx, "indexDocuments", func() (any, error) {
				return nil, a.docStore.Index(ctx, documents)
			})
			if err != nil {
				return nil, fmt.Errorf("failed to index documents: %w", err)
			}

			return nil, nil
		},
	)

	a.chatbotFlow = genkit.DefineFlow(a.g, "chatbotFlow",
		func(ctx context.Context, question string) (string, error) {
			docs, err := a.docStore.Retrieve(ctx, question, 6)
			if err != nil {
				return "", fmt.Errorf("failed to retrieve documents: %w", err)
			}

			resp, err := genkit.Generate(ctx, a.g,
				ai.WithSystem(`
You are a helpful assistant and you answer questions.
You have access to a set of documents. Use them to answer the user's question.
Don't make up answers, only use the documents provided. If you don't know the answer say it.`),
				ai.WithPrompt(question),
				ai.WithDocs(docs...),
				ai.WithOutputInstructions("Please answer in the same language as the question, and be concise."),
			)
			if err != nil {
				return "", fmt.Errorf("failed to generate response: %w", err)
			}

			return resp.Text(), nil
		},
	)

	switch controllerType {
	case CtrlTypeCmdLine:
		a.ctrl = NewCmdLineController(a.chatbotFlow)
	case CtrlTypeHTTP:
		a.ctrl = NewHTTPController(a.chatbotFlow)
	default:
		return fmt.Errorf("unsupported controller type: %v", controllerType)
	}

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

func (a *Agent) StartChatSession() error {
	if a.ctrl == nil {
		panic("StartChatSession called on a nil controller: please call Bootstrap first")
	}

	pkg.Logger.Println("Starting chat session...")
	if err := a.ctrl.Run(); err != nil {
		return fmt.Errorf("failed to start chat session: %w", err)
	}
	return nil
}
