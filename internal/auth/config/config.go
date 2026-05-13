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

	GRPCPort      string
	JWTSecret     string
	JWTExpiration string

	FileServiceAddr string
}

func Load(log *slog.Logger) *Config {
	cfg := &Config{
		DBHost:     getEnv("AUTH_DB_HOST", "localhost"),
		DBPort:     getEnv("AUTH_DB_PORT", "5432"),
		DBUser:     getEnv("AUTH_DB_USER", "postgres"),
		DBPassword: getEnv("AUTH_DB_PASSWORD", "postgres"),
		DBName:     getEnv("AUTH_DB_NAME", "auth-service-test"),
		DBSSLMode:  getEnv("AUTH_DB_SSL_MODE", "disable"),

		GRPCPort:      getEnv("AUTH_GRPC_PORT", "50051"),
		JWTSecret:     getEnv("AUTH_JWT_SECRET", "super-secret-key"),
		JWTExpiration: getEnv("AUTH_JWT_EXPIRATION", "24h"),

		FileServiceAddr: getEnv("FILE_SERVICE_ADDR", "localhost:50052"),
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
