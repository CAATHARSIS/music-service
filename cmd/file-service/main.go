package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/CAATHARSIS/music-service/internal/file/config"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	cfg := config.Load(logger)

	fmt.Println(cfg)
}
