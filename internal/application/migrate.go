package application

import (
	"fmt"

	"github.com/corradoisidoro/orders-api/internal/model"
	"gorm.io/gorm"
)

// Migrate runs all database schema migrations for the Order domain.
// Call this from main.go or a dedicated CLI command.
func Migrate(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("cannot run migrations: db is nil")
	}

	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.Order{}, &model.LineItem{}); err != nil {
			return fmt.Errorf("database migration failed: %w", err)
		}
		return nil
	})
}
