package agent

import (
	"context"
	"fmt"
	"github.com/firebase/genkit/go/plugins/localvec"
	"github.com/thomas-marquis/genkit-mistral/mistral"
	"github.com/tmc/langchaingo/textsplitter"
	"log"
	"os"
	"strings"
	"time"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

var (
	logger = log.New(os.Stdout, "goLLMan: ", log.LstdFlags|log.Lshortfile)
)

func Bootstrap(apiToken string, controllerType ControllerType) (Controller, error) {
	ctx := context.Background()
	g, err := genkit.Init(ctx,
		genkit.WithPlugins(
			mistral.NewPlugin(apiToken,
				mistral.WithRateLimiter(mistral.NewBucketCallsRateLimiter(1, 1, time.Second))),
		),
		genkit.WithDefaultModel("mistral/mistral-small"),
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize Genkit: %w", err)
	}

	chatFlow := genkit.DefineFlow(g, "chatFlow",
		func(ctx context.Context, input string) (string, error) {
			resp, err := genkit.Generate(ctx, g,
				ai.WithSystem("You are a silly assistant."),
				ai.WithPrompt(input),
			)
			if err != nil {
				return "", fmt.Errorf("failed to generate response: %w", err)
			}
			return resp.Text(), nil
		})
	var _ = chatFlow

	if err := localvec.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize local vector store: %w", err)
	}

	docStore, retriever, err := localvec.DefineRetriever(g, "pdfQA",
		localvec.Config{Embedder: genkit.LookupEmbedder(g, "mistral", "mistral-embed")},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to define docStore: %w", err)
	}

	splitter := textsplitter.NewRecursiveCharacter(
		textsplitter.WithChunkSize(100),
		textsplitter.WithChunkOverlap(10),
	)

	indexerFlow := genkit.DefineFlow(
		g, "indexerFlow",
		func(ctx context.Context, path string) (any, error) {
			allChapters, err := genkit.Run(ctx, "extract", func() ([]chapterContent, error) {
				return fetchChapters(path)
			})
			if err != nil {
				return nil, err
			}

			docs, err := genkit.Run(ctx, "chunk", func() ([]*ai.Document, error) {
				finalChunks := make([]string, 0)
				for _, chapter := range allChapters {
					chunks, err := splitter.SplitText(chapter.Content)
					if err != nil {
						return nil, err
					}
					for _, chunk := range chunks {
						chunk = "# " + chapter.Title + "\n" + strings.TrimSpace(chunk)
						if len(chunk) > 0 {
							finalChunks = append(finalChunks, chunk)
						}
					}
				}

				d := make([]*ai.Document, len(finalChunks), len(finalChunks))
				for i, chunk := range finalChunks {
					d[i] = ai.DocumentFromText(chunk, nil)
				}
				return d, nil
			})
			if err != nil {
				return nil, err
			}

			err = localvec.Index(ctx, docs, docStore)
			return nil, err
		},
	)

	if true {
		logger.Println("Indexing started...")
		ctx = context.Background()
		_, err = indexerFlow.Run(ctx, "documents/effectiveconcurrencyingo.epub")
		if err != nil {
			return nil, fmt.Errorf("failed to run indexer flow: %w", err)
		}
		logger.Println("Indexing flow complete")
	}

	chatbotFlow := genkit.DefineFlow(g, "chatbotFlow",
		func(ctx context.Context, question string) (string, error) {
			retrieverResponse, err := ai.Retrieve(ctx, retriever, ai.WithTextDocs(question))
			if err != nil {
				return "", fmt.Errorf("failed to retrieve documents: %w", err)
			}

			resp, err := genkit.Generate(ctx, g,
				ai.WithSystem(`
You are a helpful assistant and you answer questions.
You have access to a set of documents. Use them to answer the user's question.
Don't make up answers, only use the documents provided. If you don't know the answer say it.`),
				ai.WithPrompt(question),
				ai.WithDocs(retrieverResponse.Documents...),
			)
			if err != nil {
				return "", fmt.Errorf("failed to generate response: %w", err)
			}

			return resp.Text(), nil
		},
	)

	logger.Println("Starting application...")
	switch controllerType {
	case CtrlTypeCmdLine:
		return NewCmdLineController(chatbotFlow), nil
	case CtrlTypeHTTP:
		return NewHTTPController(chatbotFlow), nil
	}

	panic("invalid controller type: " + string(controllerType))
}
