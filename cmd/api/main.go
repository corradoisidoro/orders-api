package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/corradoisidoro/orders-api/internal/application"
)

func main() {
	// CLI command: go run main.go migrate
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		runMigrations()
		return
	}

	// Load configuration
	cfg, err := application.LoadConfig()
	if err != nil {
		fmt.Println("failed to load config:", err)
		os.Exit(1)
	}

	// Connect to database
	db, err := application.ConnectDatabase(cfg)
	if err != nil {
		fmt.Println("failed to connect to database:", err)
		os.Exit(1)
	}

	// Build application with injected dependencies
	app := application.New(cfg, db)

	// Graceful shutdown context
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// Start server
	if err := app.Start(ctx); err != nil {
		fmt.Println("failed to start app:", err)
	}
}

func runMigrations() {
	cfg, err := application.LoadConfig()
	if err != nil {
		fmt.Println("failed to load config:", err)
		os.Exit(1)
	}

	db, err := application.ConnectDatabase(cfg)
	if err != nil {
		fmt.Println("failed to connect to database:", err)
		os.Exit(1)
	}

	if err := application.Migrate(db); err != nil {
		fmt.Println("migration failed:", err)
		os.Exit(1)
	}

	fmt.Println("Database migrated successfully")
}
