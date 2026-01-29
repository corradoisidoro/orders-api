package infrastructure_test

import (
	"testing"

	"github.com/corradoisidoro/orders-api/internal/infrastructure"
	"github.com/corradoisidoro/orders-api/internal/model"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

//
// --- Test: db == nil ---
//

func TestMigrate_DBIsNil(t *testing.T) {
	err := infrastructure.Migrate(nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db is nil")
}

//
// --- Test: AutoMigrate fails using a closed DB connection ---
//

func TestMigrate_AutoMigrateFails(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	sqlDB, err := db.DB()
	assert.NoError(t, err)
	sqlDB.Close()

	err = infrastructure.Migrate(db)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "migrate: auto-migrate failed")
}

//
// --- Test: Successful migration ---
//

func TestMigrate_Success(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	err = infrastructure.Migrate(db)
	assert.NoError(t, err)

	assert.True(t, db.Migrator().HasTable(&model.Order{}))
	assert.True(t, db.Migrator().HasTable(&model.LineItem{}))
}
