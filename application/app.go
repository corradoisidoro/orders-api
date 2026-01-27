package application

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"gorm.io/gorm"
)

type App struct {
	router http.Handler
	config Config
	DB     *gorm.DB
}

func New(config Config, db *gorm.DB) *App {
	app := &App{
		config: config,
		DB:     db,
	}

	app.loadRoutes()
	return app
}

func (a *App) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", a.config.ServerPort),
		Handler: a.router,
	}

	log.Printf("Server running at http://localhost:%d\n", a.config.ServerPort)

	errCh := make(chan error, 1)

	go func() {
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("server error: %w", err)
		}
	}()

	select {
	case err := <-errCh:
		// Server crashed or failed to start
		return err

	case <-ctx.Done():
		// Graceful shutdown
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		return server.Shutdown(timeoutCtx)
	}
}

// Optional: expose router for testing
func (a *App) Router() http.Handler {
	return a.router
}
