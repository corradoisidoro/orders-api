package application

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/corradoisidoro/orders-api/internal/handler"
	"github.com/corradoisidoro/orders-api/internal/repository"
	"gorm.io/gorm"
)

// HTTPServer abstracts http.Server for testability.
type HTTPServer interface {
	ListenAndServe() error
	Shutdown(ctx context.Context) error
}

type App struct {
	router       http.Handler
	config       Config
	DB           *gorm.DB
	orderHandler *handler.OrderHandler

	// ServerFactory allows injecting a fake server in tests.
	ServerFactory func(addr string, handler http.Handler) HTTPServer
}

func New(config Config, db *gorm.DB) *App {
	orderRepo := repository.NewOrderRepo(db)

	app := &App{
		config: config,
		DB:     db,
		orderHandler: &handler.OrderHandler{
			Repo: orderRepo,
		},

		ServerFactory: func(addr string, h http.Handler) HTTPServer {
			return &http.Server{
				Addr:    addr,
				Handler: h,
			}
		},
	}

	app.loadRoutes()
	return app
}

func (a *App) Start(ctx context.Context) error {
	if a.ServerFactory == nil {
		return fmt.Errorf("ServerFactory is nil")
	}

	server := a.ServerFactory(fmt.Sprintf(":%d", a.config.ServerPort), a.router)

	log.Printf("Server running at http://localhost:%d\n", a.config.ServerPort)

	errCh := make(chan error, 1)

	// Run server in background and capture non-shutdown errors.
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("server error: %w", err)
		}
	}()

	select {
	case err := <-errCh:
		// Server failed unexpectedly.
		return err

	case <-ctx.Done():
		// Graceful shutdown on context cancellation.
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		log.Println("Shutting down server...")
		if err := server.Shutdown(timeoutCtx); err != nil {
			log.Printf("shutdown error: %v", err)
			return err
		}

		return nil
	}
}

func (a *App) Router() http.Handler {
	return a.router
}
