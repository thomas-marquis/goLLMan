package migrations

import (
	_ "github.com/pgvector/pgvector-go"
	"github.com/thomas-marquis/goLLMan/internal/infrastructure/orm"
	"gorm.io/gorm"
)

// MigrateUp performs database migrations to create required tables
func MigrateUp(db *gorm.DB) error {
	// Enable pgvector extension
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS vector;").Error; err != nil {
		return err
	}

	migrator := db.Migrator()

	if !migrator.HasTable(&orm.Book{}) {
		if err := migrator.CreateTable(&orm.BookIndex{}); err != nil {
			return err
		}
	}

	if !migrator.HasTable(&orm.BookIndex{}) {
		if err := migrator.CreateTable(&orm.Book{}); err != nil {
			return err
		}
	}

	// Create book_index table with vector support
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS book_index (
			id SERIAL PRIMARY KEY,
			content TEXT NOT NULL,
			book_id INTEGER NOT NULL,
			embedding vector(1024) NOT NULL,
			FOREIGN KEY (book_id) REFERENCES books(id)
		);
	`).Error; err != nil {
		return err
	}

	return nil
}

// MigrateDown rolls back migrations
func MigrateDown(db *gorm.DB) error {
	// Drop book_index table first due to foreign key constraints
	if err := db.Exec("DROP TABLE IF EXISTS book_index;").Error; err != nil {
		return err
	}

	// Drop books table
	if err := db.Exec("DROP TABLE IF EXISTS books;").Error; err != nil {
		return err
	}

	return nil
}
