package database

import (
	"fmt"
	"log/slog"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func NewPostgresDB(log *slog.Logger, databaseURL string) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", databaseURL)
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
