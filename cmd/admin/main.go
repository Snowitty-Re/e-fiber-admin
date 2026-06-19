package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	entClient, db, err := database.NewEntClient(cfg.Postgres)
	if err != nil {
		slog.Error("connect postgres failed", "err", err)
		os.Exit(1)
	}
	defer entClient.Close()
	defer db.Close()

	redisClient, err := database.NewRedisClient(cfg.Redis)
	if err != nil {
		slog.Error("connect redis failed", "err", err)
		os.Exit(1)
	}
	defer redisClient.Close()

	minioClient, err := database.NewMinIOClient(cfg.MinIO)
	if err != nil {
		slog.Error("connect minio failed", "err", err)
		os.Exit(1)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := database.EnsureBucket(ctx, minioClient, cfg.MinIO.Bucket); err != nil {
		slog.Error("ensure bucket failed", "err", err)
		os.Exit(1)
	}
	cancel()

	app := fiberapp.NewApp(fiberapp.Deps{
		Config:      cfg,
		EntClient:   entClient,
		DB:          db,
		RedisClient: redisClient,
		MinIOClient: minioClient,
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
