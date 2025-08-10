package agent

import (
	"context"
	"errors"
	"fmt"
	"github.com/firebase/genkit/go/ai"
	"github.com/thomas-marquis/goLLMan/internal/domain"
	"github.com/thomas-marquis/goLLMan/pkg"
	"math/rand"
	"sync"
)

type dataToInsert struct {
	Content   string
	Embedding []float32
	Index     int
	Batch     int
}

const numInsertWorkers = 10

func indexDocuments(bookVectorStore domain.BookVectorStore, ctx context.Context, book domain.Book, docs []*ai.Document) error {
	splitter := makeTextSplitter()

	preparedDocs, err := splitDocuments(splitter, docs)
	if err != nil {
		return fmt.Errorf("failed to split documents: %w", err)
	}

	documentsBatches := batchDocuments(preparedDocs, 50)
	pkg.Logger.Printf("Indexing %d documents for book %s through %d batches",
		len(preparedDocs), book.Title, len(documentsBatches))

	done := make(chan struct{})
	cancel := make(chan struct{})
	defer close(cancel)
	go func(cancel <-chan struct{}, done chan<- struct{}) {
		once := sync.Once{}
		for range cancel {
			once.Do(func() {
				close(done)
			})
		}
	}(cancel, done)

	data := make(chan dataToInsert)
	defer close(data)

	errCh := make(chan error)
	defer close(errCh)

	var wg sync.WaitGroup
	wg.Add(numInsertWorkers)
	for i := 0; i < numInsertWorkers; i++ {
		go func(data <-chan dataToInsert, done <-chan struct{}, errs chan<- error, cancel chan<- struct{}) {
			for {
				select {
				case <-done:
					wg.Done()
					return
				case d := <-data:
					if err := bookVectorStore.Index(ctx, book, d.Content, d.Embedding); err != nil {
						errs <- fmt.Errorf("failed to index document with content at index %d in batch %d: %w",
							d.Index, d.Batch, err)
						cancel <- struct{}{}
					}
				}
			}
		}(data, done, errCh, cancel)
	}

	errs := make([]error, 0)
	var errWg sync.WaitGroup
	errWg.Add(1)
	go func() {
		defer errWg.Done()
		for err := range errCh {
			errs = append(errs, err)
		}
	}()

	for i, batch := range documentsBatches {
		batchSize := len(batch)
		if batchSize == 0 {
			continue
		}

		//embedRes, err := ai.Embed(ctx, s.embedder, ai.WithDocs(batch...))
		embedRes, err := getFakeEmbeddingResult(batchSize)
		if err != nil {
			return fmt.Errorf("failed to embed documents: %w", err)
		}

		for j, doc := range batch {
			data <- dataToInsert{
				Content:   pkg.ContentToText(doc.Content),
				Embedding: embedRes.Embeddings[j].Embedding,
				Index:     j,
				Batch:     i,
			}
		}

		pkg.Logger.Printf("Batch no %d/%d indexed with %d documents for book %s",
			i+1, len(documentsBatches), batchSize, book.Title)
	}

	close(done)
	wg.Wait()
	errWg.Wait()

	if len(errs) > 0 {
		return fmt.Errorf("failed to index documents: %w", errors.Join(errs...))
	}

	return nil
}

type fakeEmbedding struct {
	Embedding []float32
}

type fakeEmbeddingResult struct {
	Embeddings []fakeEmbedding
}

func getFakeEmbeddingResult(size int) (fakeEmbeddingResult, error) {
	embeddings := make([]fakeEmbedding, size)
	for i := range embeddings {
		embedding := make([]float32, 1024)
		for j := range embedding {
			embedding[j] = rand.Float32()
		}
		embeddings[i] = fakeEmbedding{
			Embedding: embedding,
		}
	}
	return fakeEmbeddingResult{
		Embeddings: embeddings,
	}, nil
}
