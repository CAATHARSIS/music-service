package main

import (
	"embed"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lib/pq"
	authpb "github.com/CAATHARSIS/music-service/api/gen/auth"
	"github.com/CAATHARSIS/music-service/internal/auth/config"
	"github.com/CAATHARSIS/music-service/internal/auth/repository"
	"github.com/CAATHARSIS/music-service/internal/auth/service"
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
		logger.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	logger.Info("database connection", "status", "ok")

	if err := migrate.Run(migrationsFS, cfg.DatabaseURL(), logger); err != nil {
		logger.Error("migrations failed", "error", err)
		os.Exit(1)
	}

	repo := repository.NewRepository(db, logger)
	srv := service.NewAuthService(repo, cfg, logger)

	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		logger.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(interceptor.Logging(logger)),
	)

	authpb.RegisterAuthServiceServer(grpcServer, srv)
	reflection.Register(grpcServer)

	logger.Info("Auth Service starting", "port", cfg.GRPCPort)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("failed to serve", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down auth service...")
	grpcServer.GracefulStop()
}
