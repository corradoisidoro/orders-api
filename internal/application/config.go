package application

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort  uint16
	DatabaseDSN string
}

func LoadConfig() (Config, error) {
	if err := godotenv.Load(); err != nil {
		return Config{}, fmt.Errorf(".env file missing or unreadable: %w", err)
	}

	cfg := Config{
		ServerPort:  3000,
		DatabaseDSN: "",
	}

	// SERVER_PORT
	if portStr := os.Getenv("SERVER_PORT"); portStr != "" {
		port, err := strconv.ParseUint(portStr, 10, 16)
		if err != nil {
			return cfg, fmt.Errorf("invalid SERVER_PORT: %w", err)
		}
		cfg.ServerPort = uint16(port)
	}

	// DATABASE_DSN
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		return cfg, fmt.Errorf("DATABASE_DSN is required but not set")
	}
	cfg.DatabaseDSN = dsn

	return cfg, nil
}
