package agent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/thomas-marquis/goLLMan/agent/session"
	"github.com/thomas-marquis/goLLMan/internal/domain"
	"github.com/thomas-marquis/goLLMan/pkg"
)

const (
	systemPrompt = `You are an assistant designed to help users by answering their questions using information from a specific book.
Some of extract of this book will be provided with the users question or message. Before responding, search the book
for relevant information. If the answer is not found in the book, politely state that the information is not available
in the current source. Make sure to use only the information from the book to formulate your responses. Be concise and
accurate with your answer. Use the same language as the user.  The book information aren't send to you by the user
himself, so don't talk him about that. Just summarize and answer the question.`
)

func (a *Agent) indexerFlowHandler(ctx context.Context, book domain.Book) (any, error) {
	if book.Metadata == nil {
		book.Metadata = make(map[string]any)
	}

	book.Status = domain.StatusIndexing
	if err := a.bookRepository.Update(ctx, book); err != nil {
		pkg.Logger.Printf("Error updating book status to indexing: %s\n", err)
	}

	time.Sleep(10 * time.Second)

	parts, err := genkit.Run(ctx, "loadDocuments", func() ([]*ai.Document, error) {
		file, err := a.fileRepository.Load(ctx, book.File)
		if err != nil {
			return nil, fmt.Errorf("failed to load book content %s (%s): %w",
				book.Title, book.Author, err)
		}

		docs, err := a.docLoader.Load(book, file)
		if err != nil {
			return nil, fmt.Errorf("failed to load book %s (%s): %w",
				book.Title, book.Author, err)
		}
		return docs, nil
	})
	if err != nil {
		book.Status = domain.StatusError
		book.Metadata["error"] = err.Error()

		if err := a.bookRepository.Update(ctx, book); err != nil {
			pkg.Logger.Printf("Error updating book status: %s\n", err)
		}
		return nil, fmt.Errorf("failed to load documents from book %s (%s): %w",
			book.Title, book.Author, err)
	}

	if _, err := genkit.Run(ctx, "indexDocuments", func() (any, error) {
		if err := indexDocuments(a.embedder, a.bookVectorStore, ctx, book, parts); err != nil {
			return nil, err
		}

		book.Status = domain.StatusIndexed
		if err := a.bookRepository.Update(ctx, book); err != nil {
			pkg.Logger.Printf("Error updating book status: %s\n", err)
		}
		return nil, nil
	}); err != nil {
		book.Status = domain.StatusError
		book.Metadata["error"] = err.Error()

		if err := a.bookRepository.Update(ctx, book); err != nil {
			pkg.Logger.Printf("Error updating book status: %s\n", err)
		}
		return nil, fmt.Errorf("failed to index documents: %w", err)
	}

	return genkit.Run(ctx, "updateBookStatus", func() (any, error) {
		book.Status = domain.StatusIndexed
		return nil, a.bookRepository.Update(ctx, book)
	})
}

func (a *Agent) chatbotAiFlowHandler(ctx context.Context, input ChatbotInput) (string, error) {
	docs, err := genkit.Run(ctx, "retrieveDocuments", func() ([]*ai.Document, error) {
		resp, err := a.retriever.Retrieve(ctx, &ai.RetrieverRequest{
			Query: ai.DocumentFromText(input.Question, map[string]any{
				"limit": a.cfg.RetrievalLimit,
			}),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve documents: %w", err)
		}
		return resp.Documents, nil
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

		return prevMsg, nil
	})
	if err != nil {
		return "", err
	}

	resp, err := genkit.Run(ctx, "generateResponse", func() (*ai.ModelResponse, error) {
		return genkit.Generate(ctx, a.g,
			ai.WithSystem(systemPrompt),
			ai.WithMessages(prevMsg...),
			ai.WithPrompt(input.Question),
			ai.WithDocs(docs...),
		)
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate response: %w", err)
	}

	return genkit.Run(ctx, "updateSessionAfter", func() (string, error) {
		assistantMsg := ai.NewMessage(resp.Message.Role, resp.Message.Metadata, resp.Message.Content...)

		if err := sess.AddMessage(assistantMsg); err != nil {
			return "", fmt.Errorf("failed to add assitant message to session: %w", err)
		}
		return resp.Text(), nil
	})
}

func (a *Agent) initSession(ctx context.Context, sessionID string) (*session.Session, error) {
	sess, err := a.sessionStore.NewSession(ctx, session.WithLimit(a.cfg.SessionMessageLimit))
	if err != nil {
		return nil, fmt.Errorf("failed to create new session: %w", err)
	}
	return sess, nil
}
