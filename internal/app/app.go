// Package app wires all dependencies together and provides the HTTP server.
package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/agribid/agribid-shudh-backend/internal/config"
	"github.com/agribid/agribid-shudh-backend/internal/middleware"
	"github.com/agribid/agribid-shudh-backend/internal/response"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// App holds the wired application and the HTTP server.
type App struct {
	Router      *chi.Mux
	Server      *http.Server
	MetricsSrv  *http.Server
	Pool        *pgxpool.Pool
	Config      *config.Config
}

// New creates a new App, wires the router, middleware and health routes.
func New(cfg *config.Config, pool *pgxpool.Pool) *App {
	r := chi.NewRouter()

	// Global middleware chain — design.md §6.2 order.
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recovery)

	a := &App{
		Router: r,
		Pool:   pool,
		Config: cfg,
	}

	// System health routes (no auth required).
	a.registerHealthRoutes()

	return a
}

// registerHealthRoutes adds /health, /ready endpoints.
func (a *App) registerHealthRoutes() {
	// Liveness — always 200.
	a.Router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Readiness — pings DB pool.
	a.Router.Get("/ready", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := a.Pool.Ping(ctx); err != nil {
			response.JSON(w, http.StatusServiceUnavailable, map[string]string{
				"status": "not_ready",
				"reason": "database unreachable",
			})
			return
		}
		response.JSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})
}

// Start starts the main HTTP server and the metrics server.
func (a *App) Start() error {
	// Metrics server on separate port.
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.Handler())
	a.MetricsSrv = &http.Server{
		Addr:    fmt.Sprintf(":%d", a.Config.MetricsPort),
		Handler: metricsMux,
	}
	go func() {
		_ = a.MetricsSrv.ListenAndServe()
	}()

	// Main API server.
	a.Server = &http.Server{
		Addr:         fmt.Sprintf(":%d", a.Config.ServerPort),
		Handler:      a.Router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return a.Server.ListenAndServe()
}

// Shutdown gracefully shuts down both servers.
func (a *App) Shutdown(ctx context.Context) error {
	if a.MetricsSrv != nil {
		_ = a.MetricsSrv.Shutdown(ctx)
	}
	if a.Server != nil {
		return a.Server.Shutdown(ctx)
	}
	return nil
}
