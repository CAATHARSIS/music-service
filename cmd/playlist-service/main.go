package main

import (
	"embed"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	authpb "github.com/CAATHARSIS/music-service/api/gen/auth"
	catalogpb "github.com/CAATHARSIS/music-service/api/gen/catalog"
	playlistpb "github.com/CAATHARSIS/music-service/api/gen/playlist"
	"github.com/CAATHARSIS/music-service/internal/playlist/config"
	"github.com/CAATHARSIS/music-service/internal/playlist/repository"
	"github.com/CAATHARSIS/music-service/internal/playlist/service"
	"github.com/CAATHARSIS/music-service/pkg/database"
	"github.com/CAATHARSIS/music-service/pkg/interceptor"
	"github.com/CAATHARSIS/music-service/pkg/migrate"
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

	catalogConn, err := grpc.NewClient(
		cfg.CatalogServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logger.Error("failed to connect to catalog service", "error", err)
		os.Exit(1)
	}
	defer catalogConn.Close()

	catalogClient := catalogpb.NewCatalogServiceClient(catalogConn)

	repo := repository.NewRepository(db, logger)
	srv := service.NewPlaylistService(repo, catalogClient, logger)

	// gRPC сервер
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
		grpc.ChainUnaryInterceptor(interceptor.Logging(logger)),
		grpc.ChainUnaryInterceptor(interceptor.Auth(authClient, logger)),
	)

	playlistpb.RegisterPlaylistServiceServer(grpcServer, srv)
	reflection.Register(grpcServer)

	logger.Info("Playlist Service starting", "port", cfg.GRPCPort)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("failed to serve", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down...")
	grpcServer.GracefulStop()
}
