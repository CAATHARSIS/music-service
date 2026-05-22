// internal/ruleengine/config/config.go
package config

import (
	"fmt"
	"log/slog"
	"os"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	GRPCPort string

	AuthServiceAddr     string
	CatalogServiceAddr  string
	PlaylistServiceAddr string
}

func Load(log *slog.Logger) *Config {
	cfg := &Config{
		DBHost:     getEnv("RULES_DB_HOST", "localhost"),
		DBPort:     getEnv("RULES_DB_PORT", "5432"),
		DBUser:     getEnv("RULES_DB_USER", "postgres"),
		DBPassword: getEnv("RULES_DB_PASSWORD", "postgres"),
		DBName:     getEnv("RULES_DB_NAME", "rules_db"),
		DBSSLMode:  getEnv("RULES_DB_SSL_MODE", "disable"),

		GRPCPort: getEnv("RULES_GRPC_PORT", "50055"),

		AuthServiceAddr:     getEnv("AUTH_SERVICE_ADDR", "localhost:50051"),
		CatalogServiceAddr:  getEnv("CATALOG_SERVICE_ADDR", "localhost:50053"),
		PlaylistServiceAddr: getEnv("PLAYLIST_SERVICE_ADDR", "localhost:50054"),
	}

	log.Info("configuration loaded",
		"db_host", cfg.DBHost,
		"db_name", cfg.DBName,
		"grpc_port", cfg.GRPCPort,
	)

	return cfg
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultValue
}

func (c *Config) DatabaseURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.DBUser,
		c.DBPassword,
		c.DBHost,
		c.DBPort,
		c.DBName,
		c.DBSSLMode,
	)
}
