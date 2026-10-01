package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WithTx begins a transaction on pool, passes it to fn, and then either
// commits (when fn returns nil) or rolls back (when fn returns a non-nil
// error).  The caller receives any error from Begin, fn itself, or Commit.
func WithTx(ctx context.Context, pool *pgxpool.Pool, fn func(pgx.Tx) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		// Best-effort rollback; the original error from fn takes precedence.
		_ = tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}
