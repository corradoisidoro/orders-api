package infrastructure_test

import (
	"testing"

	"github.com/corradoisidoro/orders-api/internal/application"
	"github.com/corradoisidoro/orders-api/internal/infrastructure"
	"github.com/stretchr/testify/assert"
)

//
// --- Test: Missing DSN ---
//

func TestConnectDatabase_MissingDSN(t *testing.T) {
	cfg := application.Config{
		DatabaseDSN: "",
	}

	db, err := infrastructure.ConnectDatabase(cfg)

	assert.Nil(t, db)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database: missing DSN")
}

//
// --- Test: Invalid DSN (connection failure) ---
//

func TestConnectDatabase_InvalidDSN(t *testing.T) {
	cfg := application.Config{
		DatabaseDSN: "invalid-dsn",
	}

	db, err := infrastructure.ConnectDatabase(cfg)

	assert.Nil(t, db)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database: connect failed")
}
