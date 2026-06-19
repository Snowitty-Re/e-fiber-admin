package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/Snowitty-Re/e-fiber-admin/internal/config"
	"github.com/Snowitty-Re/e-fiber-admin/internal/database"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		logger.Error("load config failed", "err", err)
		os.Exit(1)
	}

	client, _, err := database.NewEntClient(cfg.Postgres)
	if err != nil {
		logger.Error("connect postgres failed", "err", err)
		os.Exit(1)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := client.Schema.Create(ctx); err != nil {
		logger.Error("apply schema migration failed", "err", err)
		os.Exit(1)
	}
	logger.Info("schema migration applied successfully")

	if err := database.Seed(ctx, client, cfg.Seed); err != nil {
		logger.Error("seed failed", "err", err)
		os.Exit(1)
	}
}
