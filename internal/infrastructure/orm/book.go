package orm

import (
	"encoding/json"
	"fmt"
	"github.com/thomas-marquis/goLLMan/internal/domain"
	"github.com/thomas-marquis/goLLMan/pkg"
	"gorm.io/datatypes"
)

// Book represents the ORM entity for books table
type Book struct {
	ID       uint           `gorm:"primaryKey;autoIncrement"`
	Title    string         `gorm:"not null"`
	Author   string         `gorm:"not null"`
	Metadata datatypes.JSON `gorm:"type:jsonb"`
}

// ToDomain converts the ORM entity to domain entity
func (b Book) ToDomain() domain.Book {
	var metadata map[string]any
	if b.Metadata != nil {
		var buff []byte
		if err := b.Metadata.UnmarshalJSON(buff); err != nil {
			pkg.Logger.Printf("Failed to unmarshal metadata from the gorm book entity: %s\n", err.Error())
			metadata = nil
		}
		if err := json.Unmarshal(buff, &metadata); err != nil {
			pkg.Logger.Printf("Failed to unmarshal metadata from the stored json in book: %s\n", err.Error())
			metadata = nil
		}
	}

	return domain.Book{
		ID:       idToString(b.ID),
		Title:    b.Title,
		Author:   b.Author,
		Metadata: metadata,
	}
}

// BookFromDomain creates an ORM entity from domain entity
func BookFromDomain(book domain.Book) (*Book, error) {
	var metadata datatypes.JSON
	var err error

	if book.Metadata != nil {
		metadata, err = json.Marshal(book.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	return &Book{
		Title:    book.Title,
		Author:   book.Author,
		Metadata: metadata,
	}, nil
}

// Helper to convert uint ID to string
func idToString(id uint) string {
	return fmt.Sprintf("%d", id)
}
