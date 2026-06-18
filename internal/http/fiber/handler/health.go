package handler

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"

	"github.com/Snowitty/e-fiber-admin/internal/ent"
)

type HealthHandler struct {
	entClient   *ent.Client
	redisClient *redis.Client
}

func NewHealthHandler(entClient *ent.Client, redisClient *redis.Client) *HealthHandler {
	return &HealthHandler{entClient: entClient, redisClient: redisClient}
}

func (h *HealthHandler) Healthz(c fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "ok"})
}

func (h *HealthHandler) Readyz(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 3*time.Second)
	defer cancel()

	if _, err := h.entClient.Store.Query().Count(ctx); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status": "unavailable",
			"error":  "postgres unavailable",
		})
	}
	if err := h.redisClient.Ping(ctx).Err(); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status": "unavailable",
			"error":  "redis unavailable",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "ready"})
}
