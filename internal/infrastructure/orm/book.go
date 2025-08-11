package orm

import (
	"fmt"
	"github.com/thomas-marquis/goLLMan/internal/domain"
	"gorm.io/datatypes"
)

// Book represents the ORM entity for books table
type Book struct {
	ID       uint   `gorm:"primaryKey;autoIncrement"`
	Title    string `gorm:"not null"`
	Author   string `gorm:"not null"`
	Selected bool   `gorm:"not null;default:false"`
	FileName string
	Metadata datatypes.JSONMap `gorm:"type:jsonb"`
}

// ToDomain converts the ORM entity to domain entity
func (b Book) ToDomain() domain.Book {
	return domain.Book{
		ID:       idToString(b.ID),
		Title:    b.Title,
		Author:   b.Author,
		Metadata: b.Metadata,
		Selected: b.Selected,
		File: &domain.File{
			Name: b.FileName,
		},
	}
}

// BookFromDomain creates an ORM entity from domain entity
func BookFromDomain(book domain.Book) (*Book, error) {
	var fileName string
	if book.File != nil {
		fileName = book.File.Name
	}
	return &Book{
		Title:    book.Title,
		Author:   book.Author,
		Metadata: book.Metadata,
		Selected: book.Selected,
		FileName: fileName,
	}, nil
}

// Helper to convert uint ID to string
func idToString(id uint) string {
	return fmt.Sprintf("%d", id)
}
