package application

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	ServerPort  uint16
	DatabaseDSN string
}

func LoadConfig() (Config, error) {
	cfg := Config{
		ServerPort:  3000,
		DatabaseDSN: "orders.db",
	}

	// SERVER_PORT override
	if portStr := os.Getenv("SERVER_PORT"); portStr != "" {
		port, err := strconv.ParseUint(portStr, 10, 16)
		if err != nil {
			return cfg, fmt.Errorf("invalid SERVER_PORT: %w", err)
		}
		cfg.ServerPort = uint16(port)
	}

	// DATABASE_DSN override
	if dsn := os.Getenv("DATABASE_DSN"); dsn != "" {
		cfg.DatabaseDSN = dsn
	}

	return cfg, nil
}
