package main

import (
	"embed"
	"fmt"
	"log/slog"
	"os"

	"github.com/CAATHARSIS/music-service/internal/auth/config"
	"github.com/CAATHARSIS/music-service/internal/auth/database"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	cfg := config.Load(logger)

	db, err := database.NewPostgresDB(logger, cfg)
	if err != nil {
		logger.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	logger.Info("database connection", "status", "ok")

	if err := runMigrations(cfg, logger); err != nil {
		logger.Error("migrations failed", "error", err)
		os.Exit(1)
	}
}

func runMigrations(cfg *config.Config, log *slog.Logger) error {
	log.Info("preparing migrations...")

	sourceDriver, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create iofs source: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", sourceDriver, cfg.DatabaseURL())
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	log.Info("running database migrations...")

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	if err == migrate.ErrNoChange {
		log.Info("no new migrations to apply")
	} else {
		log.Info("migrations applied successfully")
	}

	return nil
}
