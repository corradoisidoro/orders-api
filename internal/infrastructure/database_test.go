package infrastructure_test

import (
	"testing"

	"github.com/corradoisidoro/orders-api/internal/application"
	"github.com/corradoisidoro/orders-api/internal/infrastructure"
	"github.com/stretchr/testify/assert"
)

func TestConnectDatabase_MissingDSN(t *testing.T) {
	cfg := application.Config{
		DatabaseDSN: "",
	}

	db, err := infrastructure.ConnectDatabase(cfg)

	assert.Nil(t, db)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database: missing DSN")
}

func TestConnectDatabase_InvalidDSN(t *testing.T) {
	cfg := application.Config{
		DatabaseDSN: "invalid-dsn",
	}

	db, err := infrastructure.ConnectDatabase(cfg)

	assert.Nil(t, db)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database: connect failed")
}
