package infrastructure

import (
	"context"
	"github.com/thomas-marquis/goLLMan/internal/domain"
	"github.com/thomas-marquis/goLLMan/pkg"
	"os"
	"strings"
)

const (
	sep = string(os.PathSeparator)
)

type FileLocalStore struct {
	dirPath string
}

func NewFileLocalStore(dirPath string) *FileLocalStore {
	dirPath = strings.TrimSuffix(dirPath, sep)

	if err := os.MkdirAll(dirPath, 0755); err != nil {
		pkg.Logger.Fatal(err)
	}

	return &FileLocalStore{
		dirPath: dirPath,
	}
}

var _ domain.FileRepository = (*FileLocalStore)(nil)

func (f *FileLocalStore) Store(ctx context.Context, file *domain.FileWithContent) error {
	filePath := f.dirPath + sep + file.Name
	_, err := os.Stat(filePath)
	if os.IsExist(err) {
		return domain.ErrFileAlreadyExists
	}

	ff, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer ff.Close()

	_, err = ff.Write(file.Content)
	if err != nil {
		return err
	}

	return nil
}

func (f *FileLocalStore) Load(ctx context.Context, file domain.File) (*domain.FileWithContent, error) {
	filePath := f.dirPath + sep + file.Name
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, domain.ErrFileNotFound
	}

	ff, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer ff.Close()

	var content []byte
	if _, err := ff.Read(content); err != nil {
		return nil, err
	}

	return &domain.FileWithContent{
		File:    file,
		Content: content,
	}, nil
}
