package orm

import "github.com/pgvector/pgvector-go"

// BookPart represents the ORM entity for book_index table
type BookPart struct {
	ID        uint            `gorm:"primaryKey;autoIncrement"`
	BookID    uint            `gorm:"not null"`
	Content   string          `gorm:"not null"`
	Embedding pgvector.Vector `gorm:"type:vector(1024);not null"`
	Book      Book            `gorm:"foreignKey:BookID"`
}
