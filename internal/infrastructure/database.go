package infrastructure

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/corradoisidoro/orders-api/internal/application"
)

func ConnectDatabase(cfg application.Config) (*gorm.DB, error) {
	if strings.TrimSpace(cfg.DatabaseDSN) == "" {
		return nil, fmt.Errorf("database: missing DSN")
	}

	db, err := gorm.Open(postgres.Open(cfg.DatabaseDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("database: connect failed: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("database: sql.DB unwrap failed: %w", err)
	}

	// Postgres connection pool
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)

	// Ping to ensure DB is reachable
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("database: ping failed: %w", err)
	}

	return db, nil
}
