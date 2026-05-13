package main

import (
	"context"
	"embed"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	filepb "github.com/CAATHARSIS/music-service/api/gen/file"
	"github.com/CAATHARSIS/music-service/internal/file/config"
	"github.com/CAATHARSIS/music-service/internal/file/repository"
	"github.com/CAATHARSIS/music-service/internal/file/service"
	"github.com/CAATHARSIS/music-service/pkg/database"
	"github.com/CAATHARSIS/music-service/pkg/interceptor"
	"github.com/CAATHARSIS/music-service/pkg/migrate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	cfg := config.Load(logger)

	db, err := database.NewPostgresDB(logger, cfg.DatabaseURL())
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := migrate.Run(migrationsFS, cfg.DatabaseURL(), logger); err != nil {
		logger.Error("migrations failed", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	minioRepo, err := repository.NewMinioRepo(
		ctx,
		cfg.S3Endpoint,
		cfg.S3AccessKey,
		cfg.S3SecretKey,
		cfg.S3Bucket,
		cfg.S3UseSSL,
		logger,
	)
	cancel()
	if err != nil {
		logger.Error("failed to connect to MinIO", "error", err)
		os.Exit(1)
	}

	pgRepo := repository.NewPostgresRepo(db, logger)
	srv := service.NewFileService(pgRepo, minioRepo, logger)

	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		logger.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(interceptor.Logging(logger)),
	)

	filepb.RegisterFileServiceServer(grpcServer, srv)
	reflection.Register(grpcServer)

	logger.Info("File Service starting", "port", cfg.GRPCPort)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("failed to serve", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down file service...")
	grpcServer.GracefulStop()
}
