package migrations

import (
	_ "github.com/pgvector/pgvector-go"
	"github.com/thomas-marquis/goLLMan/internal/infrastructure/orm"
	"gorm.io/gorm"
)

// MigrateUp performs database migrations to create required tables
func MigrateUp(db *gorm.DB) error {
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS vector;").Error; err != nil {
		return err
	}

	migrator := db.Migrator()

	if !migrator.HasTable(&orm.Book{}) {
		if err := migrator.CreateTable(&orm.Book{}); err != nil {
			return err
		}
	}

	if !migrator.HasTable(&orm.BookIndex{}) {
		if err := migrator.CreateTable(&orm.BookIndex{}); err != nil {
			return err
		}
	}

	return nil
}
