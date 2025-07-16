package agent

import (
	"context"
	"errors"
	"fmt"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/thomas-marquis/goLLMan/agent/session"
	"github.com/thomas-marquis/goLLMan/internal/domain"
	"github.com/thomas-marquis/goLLMan/pkg"
	"time"
)

const (
	systemPrompt = `
You are a helpful assistant and you answer questions.
You have access to a set of documents. Use them to answer the user's question.
Don't make up answers, only use the documents provided. If you don't know the answer say it.`
)

func (a *Agent) indexerFlowHandler(ctx context.Context, book domain.Book) (any, error) {
	parts, err := genkit.Run(ctx, "loadDocuments", func() ([]*ai.Document, error) {
		docs, err := a.docLoader.Load(book)
		if err != nil {
			return nil, fmt.Errorf("failed to load book %s (%s): %w",
				book.Title, book.Author, err)
		}
		return docs, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load documents from book %s (%s): %w",
			book.Title, book.Author, err)
	}

	_, err = genkit.Run(ctx, "indexDocuments", func() (any, error) {
		return nil, a.docStore.Index(ctx, book, parts)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to index documents: %w", err)
	}

	return nil, nil
}

func (a *Agent) chatbotAiFlowHandler(ctx context.Context, input ChatbotInput) (string, error) {
	docs, err := genkit.Run(ctx, "retrieveDocuments", func() ([]*ai.Document, error) {
		book, err := a.bookRepository.GetByID(ctx, input.BookID)
		if err != nil {
			return nil, fmt.Errorf("failed to get book by ID %s: %w", input.BookID, err)
		}

		return a.docStore.Retrieve(ctx, book, input.Question, 6)
	})
	if err != nil {
		return "", fmt.Errorf("failed to retrieve documents: %w", err)
	}

	sess, err := genkit.Run(ctx, "getSession", func() (*session.Session, error) {
		sess, err := a.sessionStore.GetByID(ctx, input.Session)
		if err != nil {
			if errors.Is(err, session.ErrSessionNotFound) {
				sess, err = a.initSession(ctx, input.Session)
				if err != nil {
					return nil, fmt.Errorf("failed to initialize session: %w", err)
				}
			} else {
				return nil, fmt.Errorf("failed to get session: %w", err)

			}
		}
		return sess, nil
	})
	if err != nil {
		return "", err
	}

	prevMsg, err := genkit.Run(ctx, "updateSessionBefore", func() ([]*ai.Message, error) {
		prevMsg := sess.GetMessages()
		pkg.ClearMessageContext(prevMsg)
		userMsg := ai.NewUserMessage(pkg.ContentFromText(input.Question)...)
		if err := sess.AddMessage(userMsg); err != nil {
			return nil, fmt.Errorf("failed to add user message to session: %w", err)
		}
		if err := a.sessionStore.Save(ctx, sess); err != nil {
			return nil, fmt.Errorf("failed to save session: %w", err)
		}

		return prevMsg, nil
	})
	if err != nil {
		return "", err
	}

	resp, err := genkit.Run(ctx, "generateResponse", func() (*ai.ModelResponse, error) {
		return genkit.Generate(ctx, a.g,
			ai.WithMessages(prevMsg...),
			ai.WithPrompt(input.Question),
			ai.WithDocs(docs...),
			ai.WithOutputInstructions("Please answer in the same language as the question, and be concise."),
		)
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate response: %w", err)
	}

	_, err = genkit.Run(ctx, "updateSessionAfter", func() (any, error) {
		assistantMsg := ai.NewMessage(resp.Message.Role, resp.Message.Metadata, resp.Message.Content...)

		if err := sess.AddMessage(assistantMsg); err != nil {
			return nil, fmt.Errorf("failed to add assitant message to session: %w", err)
		}
		if err := a.sessionStore.Save(ctx, sess); err != nil {
			return nil, fmt.Errorf("failed to save session: %w", err)
		}
		return nil, nil
	})
	if err != nil {
		return "", err
	}

	return resp.Text(), nil
}

func (a *Agent) chatbotFakeHandle(ctx context.Context, input ChatbotInput) (string, error) {
	sess, err := a.sessionStore.GetByID(ctx, input.Session)
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

	userMsg := ai.NewUserMessage(pkg.ContentFromText(input.Question)...)
	if err := sess.AddMessage(userMsg); err != nil {
		return "", fmt.Errorf("failed to add user message to session: %w", err)
	}
	if err := a.sessionStore.Save(ctx, sess); err != nil {
		return "", fmt.Errorf("failed to save session: %w", err)
	}

	time.Sleep(1 * time.Second) // Simulate processing delay
	fakeResponse := fmt.Sprintf("I agree with you when you say:\n%s", input.Question)
	fakeAiMsg := ai.NewModelMessage(pkg.ContentFromText(fakeResponse)...)
	if err := sess.AddMessage(fakeAiMsg); err != nil {
		return "", fmt.Errorf("failed to add fake AI message to session: %w", err)
	}
	if err := a.sessionStore.Save(ctx, sess); err != nil {
		return "", fmt.Errorf("failed to save session with fake AI message: %w", err)
	}

	return fakeResponse, nil
}

func (a *Agent) initSession(ctx context.Context, sessionID string) (*session.Session, error) {
	sess, err := a.sessionStore.NewSession(ctx, session.WithLimit(a.cfg.SessionMessageLimit))
	if err != nil {
		return nil, fmt.Errorf("failed to create new session: %w", err)
	}

	systemMsg := ai.NewMessage(
		ai.RoleSystem, nil,
		pkg.ContentFromText(systemPrompt)...,
	)
	sess.AddMessage(systemMsg)

	if err := a.sessionStore.Save(ctx, sess); err != nil {
		return nil, fmt.Errorf("failed to save new session: %w", err)
	}
	return sess, nil
}
