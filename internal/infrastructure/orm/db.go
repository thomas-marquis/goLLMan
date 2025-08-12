package orm

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"strconv"
	"time"
)

// NewGormDB creates a new GORM DB connection
func NewGormDB(dsn string) (*gorm.DB, error) {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
}

// Helper to convert uint ID to string
func idToString(id uint) string {
	return fmt.Sprintf("%d", id)
}

// stringToID converts a numeric string to an uint, panicking if the conversion fails.
func stringToID(id string) uint {
	val, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		panic(err)
	}
	return uint(val)
}
