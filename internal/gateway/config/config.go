package config

import (
	"log/slog"
	"os"
)

type Config struct {
	HTTPPort string

	CatalogServiceAddr  string
	AuthServiceAddr     string
	FileServiceAddr     string
	PlaylistServiceAddr string
	RulesServiceAddr    string
	JWTSecret           string
}

func Load(log *slog.Logger) *Config {
	cfg := &Config{
		HTTPPort: getEnv("GATEWAY_HTTP_PORT", "8080"),

		CatalogServiceAddr:  getEnv("CATALOG_SERVICE_ADDR", "localhost:50053"),
		AuthServiceAddr:     getEnv("AUTH_SERVICE_ADDR", "localhost:50051"),
		FileServiceAddr:     getEnv("FILE_SERVICE_ADDR", "localhost:50052"),
		PlaylistServiceAddr: getEnv("PLAYLIST_SERVICE_ADDR", "localhost:50054"),
		RulesServiceAddr:    getEnv("RULES_SERVICE_ADDR", "localhost:50055"),
		JWTSecret:           getEnv("JWT_SECRET", "super-secret-key"),
	}

	log.Info("gateway configuration loaded", "http_port", cfg.HTTPPort)
	return cfg
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultValue
}
