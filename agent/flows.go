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
)

const (
	systemPrompt = `You are a helpful assistant. A user will ask you a question or send you a message with all the documents necessary to answer and you have to answer him/her appropriately.
Follow ALL those rules:
* Don't make up answers. If you don't know the answer or you're not sure', just say "I don't know".
* Use pieces of information provided along the user's question to answer, NOTHING ELSE.
* Be concise and accurate.
* Answer in the same language as the user.
* If you're not sure, ask the user for clarification.
* Format your response in Markdown.`
)

func (a *Agent) indexerFlowHandler(ctx context.Context, book domain.Book) (any, error) {
	if book.Metadata == nil {
		book.Metadata = make(map[string]any)
	}

	book.Status = domain.StatusIndexing
	if err := a.bookRepository.Update(ctx, book); err != nil {
		pkg.Logger.Printf("Error updating book status to indexing: %s\n", err)
	}

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
		question := "# User's message:\n" + input.Question + "\n\n"
		return genkit.Generate(ctx, a.g,
			ai.WithSystem(systemPrompt),
			ai.WithMessages(prevMsg...),
			ai.WithPrompt(question),
			ai.WithDocs(docs...),
			ai.WithModelName(a.cfg.CompletionModel),
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
