package server

import (
	"context"
	"sync"

	"github.com/firebase/genkit/go/core"
	"github.com/thomas-marquis/goLLMan/internal/domain"
	"github.com/thomas-marquis/goLLMan/pkg"
)

type WorkType int

const (
	WorkTypeIndex WorkType = iota
)

type Work interface {
	Type() WorkType
}

type IndexWork struct {
	Flow *core.Flow[domain.Book, any, struct{}]
	Book domain.Book
	Ctx  context.Context
}

func (w IndexWork) Type() WorkType {
	return WorkTypeIndex
}

func StartBackgroundWorkers(
	nbWorkers int,
	workCh <-chan Work,
	done chan struct{},
) {
	cancel := make(chan struct{})

	var once sync.Once
	go func() {
		for range cancel {
			once.Do(func() {
				close(done)
			})
		}
	}()

	for i := 0; i < nbWorkers; i++ {
		go worker(i, workCh, done, cancel)
	}
}

func worker(idx int, workCh <-chan Work, done <-chan struct{}, cancel chan<- struct{}) {
	pkg.Logger.Printf("Starting background worker %d\n", idx)
	for {
		select {
		case w := <-workCh:
			switch w.Type() {
			case WorkTypeIndex:
				work := w.(IndexWork)
				pkg.Logger.Printf("Starting indexing for book %s\n", work.Book.Title)
				if _, err := work.Flow.Run(work.Ctx, work.Book); err != nil {
					pkg.Logger.Printf("Error indexing book %s: %s\n", work.Book.Title, err)
				} else {
					pkg.Logger.Printf("Indexing for book %s complete\n", work.Book.Title)
				}
			}
		case <-done:
			pkg.Logger.Printf("Stopping background worker %d\n", idx)
			return
		}
	}
}
