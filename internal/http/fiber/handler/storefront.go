package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"

	"github.com/Snowitty-Re/e-fiber-admin/internal/domain/product"
	"github.com/Snowitty-Re/e-fiber-admin/internal/ent"
	entproduct "github.com/Snowitty-Re/e-fiber-admin/internal/ent/product"
	"github.com/Snowitty-Re/e-fiber-admin/internal/http/fiber/dto"
	pkgerr "github.com/Snowitty-Re/e-fiber-admin/internal/pkg/errors"
	"github.com/Snowitty-Re/e-fiber-admin/internal/pkg/i18n"
)

type StorefrontHandler struct {
	entClient      *ent.Client
	productService *product.Service
}

func NewStorefrontHandler(entClient *ent.Client, productService *product.Service) *StorefrontHandler {
	return &StorefrontHandler{entClient: entClient, productService: productService}
}

func (h *StorefrontHandler) ListProducts(c fiber.Ctx) error {
	store, err := h.entClient.Store.Query().First(c.Context())
	if err != nil {
		return pkgerr.ErrServiceUnavailable.WithCause(err)
	}
	locale := i18n.FromRequest(c, store.DefaultLocale)
	currency := c.Get("X-Currency", store.DefaultCurrency)
	if currency == "" {
		currency = store.DefaultCurrency
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	query := c.Query("q", "")
	var items []*ent.Product
	var total int
	if query != "" {
		items, total, err = h.productService.Search(c.Context(), query, locale, page, pageSize)
	} else {
		items, total, err = h.productService.List(c.Context(), product.ProductFilter{
			Status: "published", Page: page, PageSize: pageSize,
			CategoryID: atoiOrZero(c.Query("category_id", "")),
		})
	}
	if err != nil {
		return err
	}

	var data []dto.StoreProductResponse = []dto.StoreProductResponse{}
	for _, p := range items {
		data = append(data, toStoreProductResponse(p, locale, currency))
	}
	return c.JSON(fiber.Map{
		"data": data,
		"pagination": fiber.Map{
			"page": page, "page_size": pageSize, "total": total,
			"has_more": int64(page*pageSize) < int64(total),
		},
	})
}

func (h *StorefrontHandler) GetProduct(c fiber.Ctx) error {
	store, err := h.entClient.Store.Query().First(c.Context())
	if err != nil {
		return pkgerr.ErrServiceUnavailable.WithCause(err)
	}
	locale := i18n.FromRequest(c, store.DefaultLocale)
	currency := c.Get("X-Currency", store.DefaultCurrency)
	if currency == "" {
		currency = store.DefaultCurrency
	}

	slug := c.Params("slug")
	p, err := h.productService.GetBySlug(c.Context(), slug)
	if err != nil {
		return err
	}
	if p.Status != entproduct.StatusPublished {
		return pkgerr.ErrNotFound
	}
	return c.JSON(fiber.Map{"product": toStoreProductResponse(p, locale, currency)})
}

func toStoreProductResponse(p *ent.Product, locale, currency string) dto.StoreProductResponse {
	resp := dto.StoreProductResponse{
		ID:          p.ID,
		Slug:        p.Slug,
		ProductType: string(p.ProductType),
		Locale:      locale,
		Currency:    currency,
	}

	var t *ent.ProductTranslation
	for _, tr := range p.Edges.Translations {
		if tr.Locale == locale {
			t = tr
			break
		}
	}
	if t == nil {
		for _, tr := range p.Edges.Translations {
			t = tr
			break
		}
	}
	if t != nil {
		resp.Title = t.Title
		resp.Subtitle = t.Subtitle
		resp.Description = t.Description
		resp.SeoTitle = t.SeoTitle
		resp.SeoDesc = t.SeoDesc
	}

	for _, v := range p.Edges.Variants {
		vr := dto.StoreVariantResponse{
			ID: v.ID, SKU: v.Sku, Inventory: v.Inventory,
			AllowBackorder: v.AllowBackorder,
		}
		for _, pr := range v.Edges.Prices {
			if pr.CurrencyCode == currency {
				vr.Price = pr.Amount
				vr.CompareAtPrice = pr.CompareAtAmount
				break
			}
		}
		resp.Variants = append(resp.Variants, vr)
	}

	return resp
}
