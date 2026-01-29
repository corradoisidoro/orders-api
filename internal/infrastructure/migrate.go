package infrastructure

import (
	"fmt"

	"github.com/corradoisidoro/orders-api/internal/model"
	"gorm.io/gorm"
)

// Migrate runs all database schema migrations for the Order domain.
// Call this from main.go or a dedicated CLI command.
func Migrate(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("migrate: db is nil")
	}

	if err := db.AutoMigrate(&model.Order{}, &model.LineItem{}); err != nil {
		return fmt.Errorf("migrate: auto-migrate failed: %w", err)
	}

	return nil
}
