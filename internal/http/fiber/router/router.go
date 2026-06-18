package router

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/Snowitty/e-fiber-admin/internal/http/fiber/handler"
	"github.com/Snowitty/e-fiber-admin/internal/http/fiber/middleware"
)

type Deps struct {
	HealthH     *handler.HealthHandler
	AuthH       *handler.AuthHandler
	JWTAuthFunc fiber.Handler
}

func Register(app *fiber.App, deps Deps) {
	app.Get("/healthz", deps.HealthH.Healthz)
	app.Get("/readyz", deps.HealthH.Readyz)
	app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

	app.Get("/api/v1", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"name":    "e-fiber-admin",
			"version": "v1",
			"docs":    "/api/v1/openapi.json",
		})
	})

	api := app.Group("/api/v1")
	admin := api.Group("/admin")

	auth := admin.Group("/auth")
	auth.Post("/login", deps.AuthH.Login)
	auth.Post("/refresh", deps.AuthH.Refresh)

	authProtected := auth.Use(deps.JWTAuthFunc)
	authProtected.Post("/logout", deps.AuthH.Logout)
	authProtected.Get("/me", deps.AuthH.Me)

	products := admin.Group("/products", deps.JWTAuthFunc)
	products.Get("/", middleware.RBAC("product:read"), func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"data": []any{}, "pagination": fiber.Map{"page": 1, "page_size": 20, "total": 0, "has_more": false}})
	})
}
