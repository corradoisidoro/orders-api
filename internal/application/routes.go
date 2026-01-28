package application

import (
	"net/http"

	"github.com/corradoisidoro/orders-api/internal/handler"
	"github.com/corradoisidoro/orders-api/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (a *App) loadRoutes() {
	r := chi.NewRouter()

	// Recommended middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Health check
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Domain routes
	r.Route("/orders", a.loadOrderRoutes)

	a.router = r
}

func (a *App) loadOrderRoutes(r chi.Router) {
	// Use the DB already injected into App
	orderRepo := repository.NewOrderRepo(a.DB)
	orderHandler := &handler.OrderHandler{Repo: orderRepo}

	r.Post("/", orderHandler.Create)
	r.Get("/", orderHandler.List)
	r.Get("/{id}", orderHandler.GetByID)
	r.Patch("/{id}", orderHandler.UpdateByID) // PATCH for partial update
	r.Delete("/{id}", orderHandler.DeleteByID)
}
