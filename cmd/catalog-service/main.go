package main

import (
	"fmt"
	"log/slog"
	"os"

	_ "github.com/lib/pq"

	"github.com/CAATHARSIS/music-service/internal/catalog/config"
	"github.com/CAATHARSIS/music-service/internal/catalog/database"
	"github.com/CAATHARSIS/music-service/internal/catalog/repository"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	cfg := config.Load(logger)

	db, err := database.NewPostgresDB(logger, cfg)
	if err != nil {
		os.Exit(1)
	}
	defer db.Close()

	repo := repository.NewRepository(db, logger)

	fmt.Println(repo)
}
