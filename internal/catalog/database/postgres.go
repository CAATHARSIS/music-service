package database

import (
	"fmt"
	"log/slog"

	"github.com/jmoiron/sqlx"
	"github.com/CAATHARSIS/music-service/internal/catalog/config"
)

func NewPostgresDB(log *slog.Logger, cfg *config.Config) (*sqlx.DB, error) {
	conStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode,
	)

	db, err := sqlx.Open("postgres", conStr)
	if err != nil {
		log.Info("database connection", "status", "fail", "error", err)
		return nil, fmt.Errorf("Failed to open database: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Info("database connection", "status", "fail", "error", err)
		return nil, fmt.Errorf("Failed to ping database: %v", err)
	}

	log.Info("database connection", "status", "ok")

	return db, nil
}
