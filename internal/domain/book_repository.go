package domain

import "context"

type BookRepository interface {
	List(ctx context.Context) ([]Book, error)
	ListSelected(ctx context.Context) ([]Book, error)
	Add(ctx context.Context, title, author string, metadata map[string]any) (Book, error)
	GetByID(ctx context.Context, id string) (Book, error)
	GetByTitleAndAuthor(ctx context.Context, title, author string) (Book, error)
	ReadFromFile(ctx context.Context, filePath string) (Book, error)
}
