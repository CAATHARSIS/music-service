package main

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	catalogpb "github.com/CAATHARSIS/music-service/api/gen/catalog"
	"github.com/CAATHARSIS/music-service/internal/catalog/config"
	"github.com/CAATHARSIS/music-service/internal/catalog/database"
	"github.com/CAATHARSIS/music-service/internal/catalog/repository"
	"github.com/CAATHARSIS/music-service/internal/catalog/service"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	cfg := config.Load(logger)

	catalogDB, err := database.NewPostgresDB(logger, cfg)
	if err != nil {
		logger.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer catalogDB.Close()

	if err := runMigrations(cfg, logger); err != nil {
		logger.Error("migrations failed", "error", err)
		os.Exit(1)
	}

	repo := repository.NewRepository(catalogDB, logger)
	srv := service.NewCatalogService(repo, logger)

	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		logger.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			loggingInterceptor(logger),
		),
	)

	catalogpb.RegisterCatalogServiceServer(grpcServer, srv)
	reflection.Register(grpcServer)

	logger.Info("Catalog Service starting", "port", cfg.GRPCPort)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("failed to serve", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down gracefully...")
	grpcServer.GracefulStop()
	logger.Info("Catalog Service stopped")
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

func loggingInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		resp, err := handler(ctx, req)

		duration := time.Since(start)
		logger.Info("request completed",
			"method", info.FullMethod,
			"duration", duration,
			"error", err)

		return resp, err
	}
}
