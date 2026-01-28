package application

import (
	"fmt"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func ConnectDatabase(cfg Config) (*gorm.DB, error) {
	if cfg.DatabaseDSN == "" {
		return nil, fmt.Errorf("missing database DSN in config")
	}

	db, err := gorm.Open(sqlite.Open(cfg.DatabaseDSN), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// SQLite allows only one writer; these settings prevent locking issues
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Improve concurrency for SQLite
	db.Exec("PRAGMA journal_mode = WAL;")

	return db, nil
}
