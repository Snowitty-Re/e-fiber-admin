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
	MediaH      *handler.MediaHandler
	ProductH    *handler.ProductHandler
	CMSH        *handler.CMSHandler
	SettingsH   *handler.SettingsHandler
	StorefrontH *handler.StorefrontHandler
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
	products.Get("/", middleware.RBAC("product:read"), deps.ProductH.List)
	products.Post("/", middleware.RBAC("product:write"), deps.ProductH.Create)
	products.Get("/:id", middleware.RBAC("product:read"), deps.ProductH.Get)
	products.Post("/:id/publish", middleware.RBAC("product:publish"), deps.ProductH.Publish)
	products.Post("/:id/archive", middleware.RBAC("product:archive"), deps.ProductH.Archive)
	products.Delete("/:id", middleware.RBAC("product:delete"), deps.ProductH.Delete)

	pages := adminProtected.Group("/pages")
	pages.Get("/", middleware.RBAC("cms:read"), deps.CMSH.ListPages)
	pages.Post("/", middleware.RBAC("cms:write"), deps.CMSH.CreatePage)
	pages.Get("/:id", middleware.RBAC("cms:read"), deps.CMSH.GetPage)
	pages.Post("/:id/publish", middleware.RBAC("cms:publish"), deps.CMSH.PublishPage)
	pages.Delete("/:id", middleware.RBAC("cms:delete"), deps.CMSH.DeletePage)

	blog := adminProtected.Group("/blog-posts")
	blog.Get("/", middleware.RBAC("cms:read"), deps.CMSH.ListBlogPosts)
	blog.Post("/", middleware.RBAC("cms:write"), deps.CMSH.CreateBlogPost)
	blog.Post("/:id/publish", middleware.RBAC("cms:publish"), deps.CMSH.PublishBlogPost)
	blog.Delete("/:id", middleware.RBAC("cms:delete"), deps.CMSH.DeleteBlogPost)

	storeSettings := adminProtected.Group("/store")
	storeSettings.Get("/", middleware.RBAC("settings:read"), deps.SettingsH.GetStore)
	storeSettings.Patch("/", middleware.RBAC("settings:write"), deps.SettingsH.UpdateStore)
	storeSettings.Patch("/:id/feature-flags", middleware.RBAC("settings:write"), deps.SettingsH.UpdateFeatureFlags)

	mediaGroup := adminProtected.Group("/media")
	mediaGroup.Get("/", middleware.RBAC("media:read"), deps.MediaH.List)
	mediaGroup.Post("/", middleware.RBAC("media:write"), deps.MediaH.Upload)
	mediaGroup.Get("/:id", middleware.RBAC("media:read"), deps.MediaH.Get)
	mediaGroup.Delete("/:id", middleware.RBAC("media:delete"), deps.MediaH.Delete)

	store := api.Group("/store")
	storeProducts := store.Group("/products")
	storeProducts.Get("/", deps.StorefrontH.ListProducts)
	storeProducts.Get("/:slug", deps.StorefrontH.GetProduct)

	storePages := store.Group("/pages")
	storePages.Get("/:slug", deps.CMSH.StoreGetPage)

	storeBlog := store.Group("/blog-posts")
	storeBlog.Get("/", deps.CMSH.StoreListBlogPosts)
}
