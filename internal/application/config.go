package application

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort          uint16
	DatabaseDSN         string
	RateLimitRequests   int
	RateLimitWindowSecs int
}

func LoadConfig() (Config, error) {
	_ = godotenv.Load()

	cfg := Config{
		ServerPort:          3000,
		RateLimitRequests:   10,
		RateLimitWindowSecs: 60,
	}

	// SERVER_PORT
	if portStr := os.Getenv("SERVER_PORT"); portStr != "" {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return cfg, fmt.Errorf("invalid SERVER_PORT: %w", err)
		}
		cfg.ServerPort = uint16(port)
	}

	// DATABASE_DSN (required)
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		return cfg, fmt.Errorf("DATABASE_DSN is required but not set")
	}
	cfg.DatabaseDSN = dsn

	// RATE_LIMIT_REQUESTS
	if reqStr := os.Getenv("RATE_LIMIT_REQUESTS"); reqStr != "" {
		req, err := strconv.Atoi(reqStr)
		if err != nil {
			return cfg, fmt.Errorf("invalid RATE_LIMIT_REQUESTS: %w", err)
		}
		cfg.RateLimitRequests = req
	}

	// RATE_LIMIT_WINDOW_SECONDS
	if winStr := os.Getenv("RATE_LIMIT_WINDOW_SECONDS"); winStr != "" {
		win, err := strconv.Atoi(winStr)
		if err != nil {
			return cfg, fmt.Errorf("invalid RATE_LIMIT_WINDOW_SECONDS: %w", err)
		}
		cfg.RateLimitWindowSecs = win
	}

	// Ensure sane values
	if cfg.RateLimitRequests < 1 {
		cfg.RateLimitRequests = 1
	}
	if cfg.RateLimitWindowSecs < 1 {
		cfg.RateLimitWindowSecs = 1
	}

	return cfg, nil
}
