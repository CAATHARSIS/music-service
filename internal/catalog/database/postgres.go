package database

import (
	"database/sql"
	"fmt"

	"github.com/CAATHARSIS/music-service/internal/catalog/config"
)

func NewPostgresDB(cfg *config.Config) (*sql.DB, error) {
	conStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode,
	)

	db, err := sql.Open("postgres", conStr)
	if err != nil {
		return nil, fmt.Errorf("Failed to open database: %v", err) 
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("Failed to ping database: %v", err)
	}

	return db, nil
}
