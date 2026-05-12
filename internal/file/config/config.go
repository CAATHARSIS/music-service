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

	S3Endpoint  string
	S3AccessKey string
	S3SecretKey string
	S3Bucket    string
	S3UseSSL    bool
}

func Load(log *slog.Logger) *Config {
	cfg := &Config{
		DBHost:     getEnv("CATALOG_DB_HOST", "localhost"),
		DBPort:     getEnv("CATALOG_DB_PORT", "5432"),
		DBUser:     getEnv("CATALOG_DB_USER", "postgres"),
		DBPassword: getEnv("CATALOG_DB_PASSWORD", "postgres"),
		DBName:     getEnv("CATALOG_DB_NAME", "catalog-service-test"),
		DBSSLMode:  getEnv("CATALOG_DB_SSL_MODE", "disable"),

		GRPCPort: getEnv("CATALOG_GRPC_PORT", "50053"),

		S3Endpoint:  getEnv("S3_ENDPOINT", "localhost:9000"),
		S3AccessKey: getEnv("S3_ACCESS_KEY", "minioadming"),
		S3SecretKey: getEnv("S3_SECRET_KEY", "minioadmin"),
		S3Bucket:    getEnv("S3_BUCKET", "music-files"),
		S3UseSSL:    getEnv("S3_USE_SSL", "false") == "true",
	}

	log.Info("configuarion loaded",
		"db_host", cfg.DBHost,
		"s3_endpoint", cfg.S3Endpoint,
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
