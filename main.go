package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/corradoisidoro/orders-api/application"
)

func main() {
	// Support CLI command: go run main.go migrate
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		runMigrations()
		return
	}

	// Normal application startup command: go run main.go
	app := application.New(application.LoadConfig())

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := app.Start(ctx); err != nil {
		fmt.Println("failed to start app:", err)
	}
}

func runMigrations() {
	cfg := application.LoadConfig()
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
