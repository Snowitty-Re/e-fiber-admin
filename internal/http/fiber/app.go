package fiberapp

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/redis/go-redis/v9"

	"github.com/Snowitty/e-fiber-admin/internal/config"
	"github.com/Snowitty/e-fiber-admin/internal/ent"
	"github.com/Snowitty/e-fiber-admin/internal/http/fiber/handler"
	pkgmw "github.com/Snowitty/e-fiber-admin/internal/http/fiber/middleware"
	"github.com/Snowitty/e-fiber-admin/internal/http/fiber/router"
)

type Deps struct {
	Config      *config.Config
	EntClient   *ent.Client
	RedisClient *redis.Client
}

func NewApp(deps Deps) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      deps.Config.App.Name,
		ErrorHandler: pkgmw.ErrorHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  30 * time.Second,
	})

	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(logger.New(logger.Config{
		Format: "${time} ${locals:requestid} ${status} ${method} ${path} ${latency}\n",
		Stream: os.Stdout,
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization", "Idempotency-Key", "Accept-Language", "X-Currency"},
	}))
	app.Use(limiter.New(limiter.Config{
		Max:        300,
		Expiration: time.Minute,
	}))

	healthH := handler.NewHealthHandler(deps.EntClient, deps.RedisClient)
	router.Register(app, healthH)

	return app
}

func Run(app *fiber.App, cfg *config.Config) error {
	addr := fmt.Sprintf(":%s", cfg.App.Port)
	slog.Info("starting http server", "addr", addr, "env", cfg.App.Env)
	return app.Listen(addr)
}
