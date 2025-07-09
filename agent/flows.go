package agent

import (
	"context"
	"errors"
	"fmt"
	"github.com/firebase/genkit/go/core"
	"github.com/thomas-marquis/genkit-mistral/mistral"
	"github.com/thomas-marquis/goLLMan/agent/docstore"
	"github.com/thomas-marquis/goLLMan/agent/loader"
	"github.com/thomas-marquis/goLLMan/agent/session"
	"github.com/thomas-marquis/goLLMan/pkg"
	"time"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

const (
	systemPrompt = `
You are a helpful assistant and you answer questions.
You have access to a set of documents. Use them to answer the user's question.
Don't make up answers, only use the documents provided. If you don't know the answer say it.`
)

type ChatbotInput struct {
	Question string `json:"question"`
	Session  string `json:"session,omitempty"`
}

type Agent struct {
	g           *genkit.Genkit
	indexerFlow *core.Flow[string, any, struct{}]
	chatbotFlow *core.Flow[ChatbotInput, string, struct{}]
	docLoader   loader.DocumentLoader
	docStore    docstore.DocStore
	store       session.Store
	cfg         Config
}

func New(cfg Config, store session.Store) *Agent {
	return &Agent{
		store: store,
		cfg:   cfg,
	}
}

func (a *Agent) Flow() *core.Flow[ChatbotInput, string, struct{}] {
	if a.chatbotFlow == nil {
		panic("Flow called on a nil flow: please call Bootstrap first")
	}
	return a.chatbotFlow
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
		func(ctx context.Context, input ChatbotInput) (string, error) {
			docs, err := a.docStore.Retrieve(ctx, input.Question, 6)
			if err != nil {
				return "", fmt.Errorf("failed to retrieve documents: %w", err)
			}

			sess, err := a.store.GetByID(ctx, input.Session)
			if err != nil {
				if errors.Is(err, session.ErrSessionNotFound) {
					sess, err = a.initSession(ctx, input.Session)
					if err != nil {
						return "", fmt.Errorf("failed to initialize session: %w", err)
					}
				} else {
					return "", fmt.Errorf("failed to get session: %w", err)

				}
			}

			prevMsg, err := sess.GetMessages()
			if err != nil {
				return "", fmt.Errorf("failed to get previous messages from session: %w", err)
			}

			resp, err := genkit.Generate(ctx, a.g,
				ai.WithMessages(prevMsg...),
				ai.WithPrompt(input.Question),
				ai.WithDocs(docs...),
				ai.WithOutputInstructions("Please answer in the same language as the question, and be concise."),
			)
			if err != nil {
				return "", fmt.Errorf("failed to generate response: %w", err)
			}

			userMsg := resp.Request.Messages[len(resp.Request.Messages)-1]
			assistantMsg := ai.NewMessage(resp.Message.Role, resp.Message.Metadata, resp.Message.Content...)

			if err := sess.AddMessage(userMsg); err != nil {
				return "", fmt.Errorf("failed to add user message to session: %w", err)
			}
			if err := sess.AddMessage(assistantMsg); err != nil {
				return "", fmt.Errorf("failed to add assitant message to session: %w", err)
			}
			if err := a.store.Save(ctx, sess); err != nil {
				return "", fmt.Errorf("failed to save session: %w", err)
			}

			return resp.Text(), nil
		},
	)

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

func (a *Agent) initSession(ctx context.Context, sessionID string) (*session.Session, error) {
	sess := session.New(
		session.WithID(sessionID),
		session.WithLimit(a.cfg.SessionMessageLimit),
	)
	systemMsg := ai.NewMessage(
		ai.RoleSystem, nil,
		pkg.ContentFromText(systemPrompt)...,
	)
	sess.AddMessage(systemMsg)

	if err := a.store.Save(ctx, sess); err != nil {
		return nil, fmt.Errorf("failed to save new session: %w", err)
	}
	return sess, nil
}
