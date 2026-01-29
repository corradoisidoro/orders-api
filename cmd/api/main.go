package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/corradoisidoro/orders-api/internal/application"
	"github.com/corradoisidoro/orders-api/internal/infrastructure"
)

func main() {
	cfg, err := application.LoadConfig()
	if err != nil {
		fmt.Println("failed to load config:", err)
		os.Exit(1)
	}

	db, err := infrastructure.ConnectDatabase(cfg)
	if err != nil {
		fmt.Println("failed to connect to database:", err)
		os.Exit(1)
	}

	// CLI command: go run main.go migrate
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		fmt.Println("Running migrations...")
		if err := infrastructure.Migrate(db); err != nil {
			fmt.Println("migration failed:", err)
			os.Exit(1)
		}
		fmt.Println("Database migrated successfully")
		return
	}

	// Build application with injected dependencies
	app := application.New(cfg, db)

	// Graceful shutdown context (SIGINT + SIGTERM)
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer cancel()

	// Start server
	if err := app.Start(ctx); err != nil {
		fmt.Println("failed to start app:", err)
		os.Exit(1)
	}
}
