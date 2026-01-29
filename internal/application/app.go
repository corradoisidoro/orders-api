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

type App struct {
	router       http.Handler
	config       Config
	DB           *gorm.DB
	orderHandler *handler.OrderHandler
}

func New(config Config, db *gorm.DB) *App {
	orderRepo := repository.NewOrderRepo(db)

	app := &App{
		config:       config,
		DB:           db,
		orderHandler: &handler.OrderHandler{Repo: orderRepo},
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
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("server error: %w", err)
		}
	}()

	select {
	case err := <-errCh:
		return err

	case <-ctx.Done():
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		log.Println("Shutting down server...")
		return server.Shutdown(timeoutCtx)
	}
}

func (a *App) Router() http.Handler {
	return a.router
}
