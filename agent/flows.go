package agent

import (
	"context"
	"errors"
	"fmt"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/thomas-marquis/goLLMan/agent/session"
	"github.com/thomas-marquis/goLLMan/pkg"
)

const (
	systemPrompt = `
You are a helpful assistant and you answer questions.
You have access to a set of documents. Use them to answer the user's question.
Don't make up answers, only use the documents provided. If you don't know the answer say it.`
)

func (a *Agent) indexerFlowHandler(ctx context.Context, path string) (any, error) {
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
}

func (a *Agent) chatbotAiFlowHandler(ctx context.Context, input ChatbotInput) (string, error) {
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

	prevMsg := sess.GetMessages()
	clearMessageContext(prevMsg)
	userMsg := ai.NewUserMessage(pkg.ContentFromText(input.Question)...)
	if err := sess.AddMessage(userMsg); err != nil {
		return "", fmt.Errorf("failed to add user message to session: %w", err)
	}
	if err := a.store.Save(ctx, sess); err != nil {
		return "", fmt.Errorf("failed to save session: %w", err)
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

	assistantMsg := ai.NewMessage(resp.Message.Role, resp.Message.Metadata, resp.Message.Content...)

	if err := sess.AddMessage(assistantMsg); err != nil {
		return "", fmt.Errorf("failed to add assitant message to session: %w", err)
	}
	if err := a.store.Save(ctx, sess); err != nil {
		return "", fmt.Errorf("failed to save session: %w", err)
	}

	return resp.Text(), nil
}

func (a *Agent) chatbotFakeHandle(ctx context.Context, input ChatbotInput) (string, error) {
	return fmt.Sprintf("I agree with you when you say:\n<< %s >>", input.Question), nil
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

func clearMessageContext(msg []*ai.Message) {
	for _, m := range msg {
		if m.Role != ai.RoleUser {
			continue
		}

		var ctxPartIdx int = -1
		for i, part := range m.Content {
			if val, exists := part.Metadata["purpose"]; exists && val == "context" {
				ctxPartIdx = i
				break
			}
		}

		if ctxPartIdx != -1 {
			m.Content = m.Content[:ctxPartIdx]
		}
	}
}
