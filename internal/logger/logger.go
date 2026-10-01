// Package logger provides a structured logging singleton backed by go.uber.org/zap.
//
// Usage:
//
//	// In main.go, once:
//	if err := logger.Init(cfg.LogLevel, cfg.Env); err != nil {
//	    log.Fatalf("logger init: %v", err)
//	}
//
//	// Everywhere else:
//	logger.Get().Info("server started", zap.Int("port", cfg.ServerPort))
package logger

import (
	"errors"
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// instance is the package-level singleton. It is set exactly once by Init.
var instance *zap.Logger

// Init creates the zap logger and stores it in the package-level singleton.
// It must be called once before any call to Get().
//
//   - level is the minimum log level string (e.g., "debug", "info", "warn", "error").
//     If empty or unrecognised it defaults to "info".
//   - env controls the output format: "production" → JSON encoder; anything else →
//     the coloured console (development) encoder.
func Init(level, env string) error {
	var lvl zapcore.Level
	if err := lvl.UnmarshalText([]byte(level)); err != nil {
		// Default to info when an unrecognised level string is supplied.
		lvl = zapcore.InfoLevel
	}

	var cfg zap.Config
	if env == "production" {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
	}

	cfg.Level = zap.NewAtomicLevelAt(lvl)

	l, err := cfg.Build()
	if err != nil {
		return fmt.Errorf("logger.Init: build zap logger: %w", err)
	}

	instance = l
	return nil
}

// Get returns the initialised *zap.Logger singleton.
// It panics if Init has not been called yet, signalling a programming error
// (logger must always be available before any other component starts).
func Get() *zap.Logger {
	if instance == nil {
		panic(errors.New("logger.Get called before logger.Init — call logger.Init in main"))
	}
	return instance
}
