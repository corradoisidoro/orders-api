package application_test

import (
	"testing"

	"github.com/corradoisidoro/orders-api/internal/application"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//
// --- Helpers ---
//

func setEnv(t *testing.T, key, value string) {
	t.Helper()
	t.Setenv(key, value)
}

//
// --- TESTS ---
//

func TestLoadConfig_DefaultsApplied(t *testing.T) {
	setEnv(t, "DATABASE_DSN", "postgres://user:pass@localhost/db")

	cfg, err := application.LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, uint16(3000), cfg.ServerPort)
	assert.Equal(t, 10, cfg.RateLimitRequests)
	assert.Equal(t, 60, cfg.RateLimitWindowSecs)
	assert.Equal(t, "postgres://user:pass@localhost/db", cfg.DatabaseDSN)
}

func TestLoadConfig_CustomValues(t *testing.T) {
	setEnv(t, "SERVER_PORT", "8000")
	setEnv(t, "DATABASE_DSN", "dsn-value")
	setEnv(t, "RATE_LIMIT_REQUESTS", "50")
	setEnv(t, "RATE_LIMIT_WINDOW_SECONDS", "120")

	cfg, err := application.LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, uint16(8000), cfg.ServerPort)
	assert.Equal(t, "dsn-value", cfg.DatabaseDSN)
	assert.Equal(t, 50, cfg.RateLimitRequests)
	assert.Equal(t, 120, cfg.RateLimitWindowSecs)
}

func TestLoadConfig_MissingDSN(t *testing.T) {
	// No DATABASE_DSN set
	_, err := application.LoadConfig()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "DATABASE_DSN is required")
}

func TestLoadConfig_InvalidPort(t *testing.T) {
	setEnv(t, "DATABASE_DSN", "x")
	setEnv(t, "SERVER_PORT", "not-a-number")

	_, err := application.LoadConfig()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid SERVER_PORT")
}

func TestLoadConfig_InvalidRateLimitRequests(t *testing.T) {
	setEnv(t, "DATABASE_DSN", "x")
	setEnv(t, "RATE_LIMIT_REQUESTS", "bad")

	_, err := application.LoadConfig()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid RATE_LIMIT_REQUESTS")
}

func TestLoadConfig_InvalidRateLimitWindow(t *testing.T) {
	setEnv(t, "DATABASE_DSN", "x")
	setEnv(t, "RATE_LIMIT_WINDOW_SECONDS", "bad")

	_, err := application.LoadConfig()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid RATE_LIMIT_WINDOW_SECONDS")
}

func TestLoadConfig_ClampsLowValues(t *testing.T) {
	setEnv(t, "DATABASE_DSN", "x")
	setEnv(t, "RATE_LIMIT_REQUESTS", "0")
	setEnv(t, "RATE_LIMIT_WINDOW_SECONDS", "-5")

	cfg, err := application.LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, 1, cfg.RateLimitRequests)
	assert.Equal(t, 1, cfg.RateLimitWindowSecs)
}
