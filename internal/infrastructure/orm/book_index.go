package orm

import "github.com/pgvector/pgvector-go"

// BookIndex represents the ORM entity for book_index table
type BookIndex struct {
	ID        uint            `gorm:"primaryKey;autoIncrement"`
	Content   string          `gorm:"not null"`
	BookID    uint            `gorm:"not null"`
	Embedding pgvector.Vector `gorm:"type:vector(1024);not null"`
	Book      Book            `gorm:"foreignKey:BookID"`
}
