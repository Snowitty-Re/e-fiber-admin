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
	"github.com/Snowitty-Re/e-fiber-admin/internal/events"
	"github.com/Snowitty-Re/e-fiber-admin/internal/jobs"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config failed", "err", err)
		os.Exit(1)
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
	bctx, bcancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := database.EnsureBucket(bctx, minioClient, cfg.MinIO.Bucket); err != nil {
		slog.Error("ensure bucket failed", "err", err)
		os.Exit(1)
	}
	bcancel()

	jobsClient := jobs.NewClient(cfg.Redis)
	subscriber := events.NewSubscriber(redisClient)
	jobsServer := jobs.NewServer(cfg.Asynq, cfg.Redis)

	seedCtx, seedCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer seedCancel()
	if err := SeedNotifications(seedCtx, entClient); err != nil {
		slog.Warn("seed notification templates skipped", "err", err)
	}
	if err := SeedOrderTemplates(seedCtx, entClient); err != nil {
		slog.Warn("seed order templates skipped", "err", err)
	}

	RegisterEventHandlers(subscriber, entClient, jobsClient)
	RegisterJobHandlers(jobsServer, entClient)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := subscriber.Run(ctx); err != nil && err != context.Canceled {
			slog.Error("subscriber stopped", "err", err)
		}
	}()

	go func() {
		if err := jobsServer.Run(); err != nil {
			slog.Error("asynq server stopped", "err", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("worker shutting down")
	cancel()
	jobsServer.Shutdown()
}
