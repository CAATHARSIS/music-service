package main

import (
	"embed"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	authpb "github.com/CAATHARSIS/music-service/api/gen/auth"
	catalogpb "github.com/CAATHARSIS/music-service/api/gen/catalog"
	filepb "github.com/CAATHARSIS/music-service/api/gen/file"
	"github.com/CAATHARSIS/music-service/internal/catalog/config"
	"github.com/CAATHARSIS/music-service/internal/catalog/repository"
	"github.com/CAATHARSIS/music-service/internal/catalog/service"
	"github.com/CAATHARSIS/music-service/pkg/database"
	"github.com/CAATHARSIS/music-service/pkg/interceptor"
	"github.com/CAATHARSIS/music-service/pkg/migrate"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	cfg := config.Load(logger)

	catalogDB, err := database.NewPostgresDB(logger, cfg.DatabaseURL())
	if err != nil {
		logger.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer catalogDB.Close()

	if err := migrate.Run(migrationsFS, cfg.DatabaseURL(), logger); err != nil {
		logger.Error("migrations failed", "error", err)
		os.Exit(1)
	}

	fileConn, err := grpc.NewClient(
		cfg.FileServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logger.Error("failed to connect to file service", "error", err)
		os.Exit(1)
	}
	defer fileConn.Close()

	fileClient := filepb.NewFileServiceClient(fileConn)

	repo := repository.NewRepository(catalogDB, logger)
	srv := service.NewCatalogService(repo, fileClient, logger)

	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		logger.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	authConn, err := grpc.NewClient(
		cfg.AuthServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logger.Error("failed to connect to auth service", "error", err)
		os.Exit(1)
	}
	defer authConn.Close()

	authClient := authpb.NewAuthServiceClient(authConn)

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptor.Logging(logger),
			interceptor.Auth(authClient, logger),
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
