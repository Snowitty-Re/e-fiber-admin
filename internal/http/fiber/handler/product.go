package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"

	"github.com/Snowitty-Re/e-fiber-admin/internal/domain/product"
	"github.com/Snowitty-Re/e-fiber-admin/internal/ent"
	"github.com/Snowitty-Re/e-fiber-admin/internal/http/fiber/dto"
	pkgerr "github.com/Snowitty-Re/e-fiber-admin/internal/pkg/errors"
)

type ProductHandler struct {
	productService *product.Service
}

func NewProductHandler(productService *product.Service) *ProductHandler {
	return &ProductHandler{productService: productService}
}

func (h *ProductHandler) Create(c fiber.Ctx) error {
	var req dto.CreateProductRequest
	if err := c.Bind().Body(&req); err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	if req.Slug == "" {
		return pkgerr.ErrValidation.WithDetails(pkgerr.FieldError{Field: "slug", Issue: "required"})
	}
	if len(req.Translations) == 0 {
		return pkgerr.ErrValidation.WithDetails(pkgerr.FieldError{Field: "translations", Issue: "at least one required"})
	}
	if len(req.Variants) == 0 {
		return pkgerr.ErrValidation.WithDetails(pkgerr.FieldError{Field: "variants", Issue: "at least one required"})
	}

	in := product.ProductInput{
		Slug:           req.Slug,
		ProductType:    req.ProductType,
		CategoryID:     req.CategoryID,
		WeightG:        req.WeightG,
		IsVirtual:      req.IsVirtual,
		IsDownloadable: req.IsDownloadable,
	}
	for _, t := range req.Translations {
		in.Translations = append(in.Translations, product.TranslationInput{
			Locale: t.Locale, Title: t.Title, Subtitle: t.Subtitle,
			Description: t.Description, Material: t.Material, Origin: t.Origin,
			Packing: t.Packing, SeoTitle: t.SeoTitle, SeoDesc: t.SeoDesc,
		})
	}
	for _, v := range req.Variants {
		vi := product.VariantInput{
			SKU: v.SKU, Barcode: v.Barcode, WeightG: v.WeightG,
			AllowBackorder: v.AllowBackorder, Inventory: v.Inventory, Position: v.Position,
		}
		for _, p := range v.Prices {
			vi.Prices = append(vi.Prices, product.VariantPriceInput{
				CurrencyCode: p.CurrencyCode, Amount: p.Amount, CompareAtAmount: p.CompareAtAmount,
			})
		}
		in.Variants = append(in.Variants, vi)
	}

	p, err := h.productService.Create(c.Context(), in)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"product": toProductResponse(p)})
}

func (h *ProductHandler) Get(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	p, err := h.productService.Get(c.Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"product": toProductResponse(p)})
}

func (h *ProductHandler) List(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))
	items, total, err := h.productService.List(c.Context(), product.ProductFilter{
		Status: c.Query("status", ""), ProductType: c.Query("product_type", ""),
		CategoryID: atoiOrZero(c.Query("category_id", "")),
		Page: page, PageSize: pageSize,
	})
	if err != nil {
		return err
	}
	var data []dto.ProductResponse
	for _, p := range items {
		data = append(data, toProductResponse(p))
	}
	return c.JSON(fiber.Map{
		"data": data,
		"pagination": fiber.Map{
			"page": page, "page_size": pageSize, "total": total,
			"has_more": int64(page*pageSize) < int64(total),
		},
	})
}

func (h *ProductHandler) Publish(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	if err := h.productService.Publish(c.Context(), id); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *ProductHandler) Archive(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	if err := h.productService.Archive(c.Context(), id); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *ProductHandler) Delete(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	if err := h.productService.Delete(c.Context(), id); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func atoiOrZero(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return n
}

func toProductResponse(p *ent.Product) dto.ProductResponse {
	resp := dto.ProductResponse{
		ID: p.ID, Slug: p.Slug, ProductType: string(p.ProductType),
		Status: string(p.Status), CategoryID: p.CategoryID, WeightG: p.WeightG,
		IsVirtual: p.IsVirtual, IsDownloadable: p.IsDownloadable,
	}
	for _, t := range p.Edges.Translations {
		resp.Translations = append(resp.Translations, dto.TranslationResponse{
			Locale: t.Locale, Title: t.Title, Subtitle: t.Subtitle,
			Description: t.Description, Material: t.Material, Origin: t.Origin,
			Packing: t.Packing, SeoTitle: t.SeoTitle, SeoDesc: t.SeoDesc,
		})
	}
	for _, v := range p.Edges.Variants {
		vr := dto.VariantResponse{
			ID: v.ID, SKU: v.Sku, Barcode: v.Barcode, WeightG: v.WeightG,
			AllowBackorder: v.AllowBackorder, Inventory: v.Inventory, Position: v.Position,
		}
		for _, pr := range v.Edges.Prices {
			vr.Prices = append(vr.Prices, dto.VariantPriceResponse{
				CurrencyCode: pr.CurrencyCode, Amount: pr.Amount, CompareAtAmount: pr.CompareAtAmount,
			})
		}
		resp.Variants = append(resp.Variants, vr)
	}
	return resp
}
