package db

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	// pgx stdlib is required so migrate can open a *sql.DB from DATABASE_URL.
	_ "github.com/jackc/pgx/v5/stdlib"
	"database/sql"
)

// RunMigrations applies all pending up-migrations from the migrations/
// directory (relative to the process working directory) against the given
// PostgreSQL database.
//
// The function is a no-op when the RUN_MIGRATIONS environment variable is not
// set to "true", so it is safe to wire unconditionally in main.go and control
// migration execution purely via environment.
//
// Migration source strategy:
//   The migrations/ directory lives at the project root.  Because it is NOT
//   inside the internal/db package we cannot use go:embed across the package
//   boundary without a separate embed-only file at the project root.  Instead
//   we use os.DirFS to open the directory relative to the working directory
//   (which is always the project root in production and in tests that set
//   Working directory correctly) and pass the resulting fs.FS to the iofs
//   source driver.  This keeps the code simple and avoids any embed
//   cross-boundary tricks.
func RunMigrations(_ context.Context, databaseURL string) error {
	if os.Getenv("RUN_MIGRATIONS") != "true" {
		// Migration runner is disabled; skip silently.
		return nil
	}

	// Open the migrations directory from the working directory.
	migrationsFS := os.DirFS("migrations")

	return runMigrationsFromFS(migrationsFS, databaseURL)
}

// runMigrationsFromFS is the internal implementation that accepts an fs.FS so
// it can be called from tests with an in-memory filesystem.
func runMigrationsFromFS(migrationsFS fs.FS, databaseURL string) error {
	// Build an iofs source from the provided filesystem.
	src, err := iofs.New(migrationsFS, ".")
	if err != nil {
		return fmt.Errorf("migrate: build source driver: %w", err)
	}

	// Open a *sql.DB for golang-migrate (it does not accept pgxpool).
	sqlDB, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return fmt.Errorf("migrate: open sql.DB: %w", err)
	}
	defer sqlDB.Close()

	// Build the postgres database driver for migrate.
	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("migrate: build postgres driver: %w", err)
	}

	// Construct the migrator.
	m, err := migrate.NewWithInstance("iofs", src, "postgres", driver)
	if err != nil {
		return fmt.Errorf("migrate: create migrator: %w", err)
	}

	// Apply all pending migrations.
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate: apply up: %w", err)
	}

	return nil
}
