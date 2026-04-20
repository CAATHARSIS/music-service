package config

import (
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
	FileServiceAddr     string
	PlaylistServiceAddr string
}

func Load(log *slog.Logger) *Config {
	cfg := &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "catalog-service-test"),
		DBSSLMode:  getEnv("DB_SSL_MODE", "disable"),

		GRPCPort: getEnv("GRPC_PORT", "50053"),

		AuthServiceAddr:     getEnv("AUTH_SERVICE_ADDR", "localhost:50051"),
		FileServiceAddr:     getEnv("FILE_SERVICE_ADDR", "localhost:50052"),
		PlaylistServiceAddr: getEnv("PLAYLIST_SERVICE_ADDR", "localhost:50054"),
	}

	log.Info("configuarion loaded",
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
