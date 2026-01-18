package application

import (
	"fmt"

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
	return db, nil
}
