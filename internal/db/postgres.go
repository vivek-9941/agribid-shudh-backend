package db

import (
	"context"
	"time"

	"github.com/agribid/agribid-shudh-backend/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

// New creates a new *pgxpool.Pool from the given Config.
// It parses the DatabaseURL, applies connection-pool tuning derived from cfg,
// and verifies connectivity by calling pool.Ping before returning.
func New(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	// Connection-pool sizing from config.
	poolCfg.MaxConns = int32(cfg.DBMaxOpenConns)
	poolCfg.MinConns = int32(cfg.DBMinOpenConns)

	// Connection lifetime limits to prevent stale connections.
	poolCfg.MaxConnLifetime = time.Hour
	poolCfg.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, err
	}

	// Verify the database is reachable before returning.
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}
