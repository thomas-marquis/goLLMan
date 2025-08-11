package domain

import "context"

type FileRepository interface {
	Store(ctx context.Context, file *FileWithContent) error
	Load(ctx context.Context, file File) (*FileWithContent, error)
}
