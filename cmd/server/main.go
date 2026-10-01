package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/agribid/agribid-shudh-backend/internal/app"
	"github.com/agribid/agribid-shudh-backend/internal/config"
	"github.com/agribid/agribid-shudh-backend/internal/db"
	"github.com/agribid/agribid-shudh-backend/internal/logger"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()

	// 1. Load configuration from environment variables.
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	// 2. Initialize structured logging.
	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}
	if err := logger.Init(cfg.LogLevel, env); err != nil {
		log.Fatalf("logger: %v", err)
	}
	defer logger.Get().Sync()

	logger.Get().Info("starting agribid-shudh-backend",
		zap.Int("server_port", cfg.ServerPort),
		zap.Int("metrics_port", cfg.MetricsPort),
		zap.String("env", env),
	)

	// 3. Run database migrations when RUN_MIGRATIONS=true.
	if os.Getenv("RUN_MIGRATIONS") == "true" {
		if err := db.RunMigrations(ctx, cfg.DatabaseURL); err != nil {
			logger.Get().Fatal("migrations failed", zap.Error(err))
		}
		logger.Get().Info("migrations completed successfully")
	}

	// 4. Create database connection pool.
	pool, err := db.New(ctx, cfg)
	if err != nil {
		logger.Get().Fatal("database connection failed", zap.Error(err))
	}
	defer pool.Close()
	logger.Get().Info("database connected",
		zap.Int("max_conns", cfg.DBMaxOpenConns),
		zap.Int("min_conns", cfg.DBMinOpenConns),
	)

	// 5. Wire application.
	application := app.New(cfg, pool)

	// 6. Start HTTP server in background.
	go func() {
		logger.Get().Info("HTTP server listening", zap.Int("port", cfg.ServerPort))
		if err := application.Start(); err != nil && err != http.ErrServerClosed {
			logger.Get().Fatal("server error", zap.Error(err))
		}
	}()

	// 7. Graceful shutdown: wait for SIGTERM or SIGINT.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	sig := <-quit
	logger.Get().Info("shutdown signal received", zap.String("signal", sig.String()))

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := application.Shutdown(shutdownCtx); err != nil {
		logger.Get().Error("forced shutdown", zap.Error(err))
	}

	logger.Get().Info("server stopped gracefully")
}
