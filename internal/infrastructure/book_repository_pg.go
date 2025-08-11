package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"github.com/firebase/genkit/go/ai"
	"github.com/pgvector/pgvector-go"
	"github.com/thomas-marquis/goLLMan/internal/domain"
	"github.com/thomas-marquis/goLLMan/internal/infrastructure/orm"
	"github.com/thomas-marquis/goLLMan/pkg"
	"github.com/timsims/pamphlet"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strconv"
)

const (
	bookMetaLocalEpubPathKey = "local_epub_filepath"
)

type BookRepositoryPostgres struct {
	db *gorm.DB
}

var _ domain.BookRepository = (*BookRepositoryPostgres)(nil)
var _ domain.BookVectorStore = (*BookRepositoryPostgres)(nil)

func NewBookRepositoryPostgres(db *gorm.DB) *BookRepositoryPostgres {
	return &BookRepositoryPostgres{
		db: db,
	}
}

func (r *BookRepositoryPostgres) List(ctx context.Context) ([]domain.Book, error) {
	var ormBooks []orm.Book

	result := r.db.WithContext(ctx).Find(&ormBooks)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domain.ErrBookNotFound
		}
		return nil, errors.Join(domain.ErrRepositoryError, result.Error)
	}

	books := make([]domain.Book, len(ormBooks))
	for i, ormBook := range ormBooks {
		books[i] = ormBook.ToDomain()
	}

	return books, nil
}

func (r *BookRepositoryPostgres) ListSelected(ctx context.Context) ([]domain.Book, error) {
	var ormBooks []orm.Book

	result := r.db.WithContext(ctx).Find(&ormBooks)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domain.ErrBookNotFound
		}
		return nil, errors.Join(domain.ErrRepositoryError, result.Error)
	}

	books := make([]domain.Book, len(ormBooks))
	for i, ormBook := range ormBooks {
		books[i] = ormBook.ToDomain()
	}

	return books, nil
}

func (r *BookRepositoryPostgres) Add(ctx context.Context, title, author string, file domain.File, metadata map[string]any) (domain.Book, error) {
	// Check if the book already exists
	var book domain.Book
	var err error

	exists := true
	book, err = r.GetByTitleAndAuthor(ctx, title, author)
	if err != nil {
		if errors.Is(err, domain.ErrBookNotFound) {
			exists = false
		} else {
			return domain.Book{}, errors.Join(domain.ErrRepositoryError, err)
		}
	}
	if exists {
		return book, domain.ErrBookAlreadyExists
	}

	// If the book does not exist, insert it
	domainBook := domain.Book{
		Title:    title,
		Author:   author,
		Metadata: metadata,
		File:     file,
	}

	ormBook, err := orm.BookFromDomain(domainBook)
	if err != nil {
		return domain.Book{}, errors.Join(domain.ErrRepositoryError, err)
	}

	result := r.db.WithContext(ctx).Create(ormBook)
	if result.Error != nil {
		return domain.Book{}, errors.Join(domain.ErrRepositoryError, result.Error)
	}

	return ormBook.ToDomain(), nil
}

func (r *BookRepositoryPostgres) GetByID(ctx context.Context, id string) (domain.Book, error) {
	bookId, err := strconv.Atoi(id)
	if err != nil {
		return domain.Book{}, errors.Join(domain.ErrRepositoryError, err)
	}

	var ormBook orm.Book
	result := r.db.WithContext(ctx).First(&ormBook, bookId)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return domain.Book{}, domain.ErrBookNotFound
		}
		return domain.Book{}, errors.Join(domain.ErrRepositoryError, result.Error)
	}

	return ormBook.ToDomain(), nil
}

func (r *BookRepositoryPostgres) GetByTitleAndAuthor(ctx context.Context, title, author string) (domain.Book, error) {
	var ormBook orm.Book
	result := r.db.WithContext(ctx).Where("title = ? AND author = ?", title, author).First(&ormBook)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return domain.Book{}, domain.ErrBookNotFound
		}
		return domain.Book{}, errors.Join(domain.ErrRepositoryError, result.Error)
	}

	return ormBook.ToDomain(), nil
}

func (r *BookRepositoryPostgres) ReadFromFile(ctx context.Context, file *domain.FileWithContent) (domain.Book, error) {
	parser, err := pamphlet.OpenBytes(file.Content)
	if err != nil {
		return domain.Book{}, fmt.Errorf("error opening epub file: %w", err)
	}

	book := parser.GetBook()
	if book == nil {
		return domain.Book{}, errors.New("invalid epub file")
	}

	return domain.Book{
		Title:  book.Title,
		Author: book.Author,
		File:   file.File,
	}, nil
}

func (r *BookRepositoryPostgres) Index(ctx context.Context, book domain.Book, content string, vector []float32) error {
	bi := orm.BookPart{
		Content:   content,
		Embedding: pgvector.NewVector(vector),
	}
	if err := r.db.WithContext(ctx).Create(&bi).Error; err != nil {
		return fmt.Errorf("failed to create book index: %w", err)
	}
	return nil
}

func (r *BookRepositoryPostgres) Retrieve(ctx context.Context, books []domain.Book, embedding []float32, limit int) ([]*ai.Document, error) {
	bookIDs := make([]int, len(books), len(books))
	for i, book := range books {
		id, _ := strconv.Atoi(book.ID)
		bookIDs[i] = id
	}

	var bis []orm.BookPart
	if err := r.db.
		WithContext(ctx).
		Clauses(clause.OrderBy{
			Expression: clause.Expr{
				SQL:  "embedding <=> ?",
				Vars: []any{pgvector.NewVector(embedding)},
			},
		}).
		Limit(limit).
		Where("book_id IN ?", bookIDs).
		Find(&bis).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			pkg.Logger.Println("No book index found, returning empty retriever response")
			return make([]*ai.Document, 0), nil
		}
		return nil, fmt.Errorf("failed to retrieve books indices: %w", err)
	}

	documents := make([]*ai.Document, len(bis), len(bis))
	for i, bi := range bis {
		documents[i] = ai.DocumentFromText(
			bi.Content,
			map[string]any{
				"id":      bi.ID,
				"book_id": bi.BookID,
			},
		)
	}

	return documents, nil
}
