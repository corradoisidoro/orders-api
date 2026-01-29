package application

import (
	"net/http"
	"time"

	appmw "github.com/corradoisidoro/orders-api/internal/middleware"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

func (a *App) loadRoutes() {
	r := chi.NewRouter()

	// Core middleware
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(60 * time.Second))

	// App middleware
	r.Use(appmw.RateLimitMiddleware(
		a.config.RateLimitRequests,
		a.config.RateLimitWindowSecs,
	))

	// Health check
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// API routes
	r.Route("/orders", a.loadOrderRoutes)

	a.router = r
}

func (a *App) loadOrderRoutes(r chi.Router) {
	r.Post("/", a.orderHandler.Create)
	r.Get("/", a.orderHandler.List)
	r.Get("/{id}", a.orderHandler.GetByID)
	r.Patch("/{id}", a.orderHandler.UpdateByID)
	r.Delete("/{id}", a.orderHandler.DeleteByID)
}
