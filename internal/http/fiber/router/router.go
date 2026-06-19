package router

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/Snowitty-Re/e-fiber-admin/internal/http/fiber/handler"
	"github.com/Snowitty-Re/e-fiber-admin/internal/http/fiber/middleware"
)

type Deps struct {
	HealthH     *handler.HealthHandler
	AuthH       *handler.AuthHandler
	RegionH     *handler.RegionHandler
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

	adminProtected := admin.Use(deps.JWTAuthFunc)

	regions := adminProtected.Group("/regions")
	regions.Get("/", middleware.RBAC("region:read"), deps.RegionH.ListRegions)
	regions.Post("/", middleware.RBAC("region:write"), deps.RegionH.CreateRegion)
	regions.Get("/:id", middleware.RBAC("region:read"), deps.RegionH.GetRegion)
	regions.Patch("/:id", middleware.RBAC("region:write"), deps.RegionH.UpdateRegion)
	regions.Delete("/:id", middleware.RBAC("region:write"), deps.RegionH.DeleteRegion)

	locales := adminProtected.Group("/locales")
	locales.Get("/", middleware.RBAC("region:read"), deps.RegionH.ListLocales)
	locales.Post("/", middleware.RBAC("region:write"), deps.RegionH.CreateLocale)

	currencies := adminProtected.Group("/currencies")
	currencies.Get("/", middleware.RBAC("region:read"), deps.RegionH.ListCurrencies)
	currencies.Post("/", middleware.RBAC("region:write"), deps.RegionH.CreateCurrency)

	taxRates := adminProtected.Group("/tax-rates")
	taxRates.Get("/", middleware.RBAC("region:read"), deps.RegionH.ListTaxRates)
	taxRates.Post("/", middleware.RBAC("region:write"), deps.RegionH.CreateTaxRate)
	taxRates.Delete("/:id", middleware.RBAC("region:write"), deps.RegionH.DeleteTaxRate)

	products := adminProtected.Group("/products")
	products.Get("/", middleware.RBAC("product:read"), func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"data": []any{}, "pagination": fiber.Map{"page": 1, "page_size": 20, "total": 0, "has_more": false}})
	})
}
