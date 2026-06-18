package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Snowitty-Re/e-fiber-admin/internal/config"
	"github.com/Snowitty-Re/e-fiber-admin/internal/database"
	fiberapp "github.com/Snowitty-Re/e-fiber-admin/internal/http/fiber"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config failed", "err", err)
		os.Exit(1)
	}
	if cfg.IsDev() {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	entClient, err := database.NewEntClient(cfg.Postgres)
	if err != nil {
		slog.Error("connect postgres failed", "err", err)
		os.Exit(1)
	}
	defer entClient.Close()

	redisClient, err := database.NewRedisClient(cfg.Redis)
	if err != nil {
		slog.Error("connect redis failed", "err", err)
		os.Exit(1)
	}
	defer redisClient.Close()

	app := fiberapp.NewApp(fiberapp.Deps{
		Config:      cfg,
		EntClient:   entClient,
		RedisClient: redisClient,
	})

	go func() {
		if err := fiberapp.Run(app, cfg); err != nil {
			slog.Error("http server stopped", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("shutting down")
	if err := app.Shutdown(); err != nil {
		slog.Error("shutdown failed", "err", err)
	}
}
