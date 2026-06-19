package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"

	"github.com/Snowitty-Re/e-fiber-admin/internal/domain/region"
	"github.com/Snowitty-Re/e-fiber-admin/internal/ent"
	"github.com/Snowitty-Re/e-fiber-admin/internal/http/fiber/dto"
	pkgerr "github.com/Snowitty-Re/e-fiber-admin/internal/pkg/errors"
)

type RegionHandler struct {
	regionService *region.Service
}

func NewRegionHandler(regionService *region.Service) *RegionHandler {
	return &RegionHandler{regionService: regionService}
}

func (h *RegionHandler) ListLocales(c fiber.Ctx) error {
	locales, err := h.regionService.ListLocales(c.Context())
	if err != nil {
		return err
	}
	var resp []dto.LocaleResponse
	for _, l := range locales {
		resp = append(resp, dto.LocaleResponse{
			ID: l.ID, Code: l.Code, Name: l.Name, IsActive: l.IsActive,
		})
	}
	return c.JSON(fiber.Map{"data": resp})
}

func (h *RegionHandler) CreateLocale(c fiber.Ctx) error {
	var req dto.CreateLocaleRequest
	if err := c.Bind().Body(&req); err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	if req.Code == "" || req.Name == "" {
		return pkgerr.ErrValidation.WithDetails(
			pkgerr.FieldError{Field: "code", Issue: "required"},
			pkgerr.FieldError{Field: "name", Issue: "required"},
		)
	}
	l, err := h.regionService.CreateLocale(c.Context(), req.Code, req.Name)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"locale": dto.LocaleResponse{
		ID: l.ID, Code: l.Code, Name: l.Name, IsActive: l.IsActive,
	}})
}

func (h *RegionHandler) ListCurrencies(c fiber.Ctx) error {
	currencies, err := h.regionService.ListCurrencies(c.Context())
	if err != nil {
		return err
	}
	var resp []dto.CurrencyResponse
	for _, cur := range currencies {
		resp = append(resp, dto.CurrencyResponse{
			ID: cur.ID, Code: cur.Code, Name: cur.Name, Symbol: cur.Symbol,
			Precision: cur.Precision, IsActive: cur.IsActive,
		})
	}
	return c.JSON(fiber.Map{"data": resp})
}

func (h *RegionHandler) CreateCurrency(c fiber.Ctx) error {
	var req dto.CreateCurrencyRequest
	if err := c.Bind().Body(&req); err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	if req.Code == "" || req.Name == "" {
		return pkgerr.ErrValidation.WithDetails(
			pkgerr.FieldError{Field: "code", Issue: "required"},
			pkgerr.FieldError{Field: "name", Issue: "required"},
		)
	}
	cur, err := h.regionService.CreateCurrency(c.Context(), req.Code, req.Name, req.Symbol, req.Precision)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"currency": dto.CurrencyResponse{
		ID: cur.ID, Code: cur.Code, Name: cur.Name, Symbol: cur.Symbol,
		Precision: cur.Precision, IsActive: cur.IsActive,
	}})
}

func (h *RegionHandler) ListRegions(c fiber.Ctx) error {
	regions, err := h.regionService.ListRegions(c.Context())
	if err != nil {
		return err
	}
	var resp []dto.RegionResponse
	for _, r := range regions {
		resp = append(resp, toRegionResponse(r))
	}
	return c.JSON(fiber.Map{"data": resp})
}

func (h *RegionHandler) GetRegion(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	r, err := h.regionService.GetRegion(c.Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"region": toRegionResponse(r)})
}

func (h *RegionHandler) CreateRegion(c fiber.Ctx) error {
	var req dto.CreateRegionRequest
	if err := c.Bind().Body(&req); err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	if req.Name == "" || req.Locale == "" || req.CurrencyCode == "" {
		return pkgerr.ErrValidation.WithDetails(
			pkgerr.FieldError{Field: "name", Issue: "required"},
			pkgerr.FieldError{Field: "locale", Issue: "required"},
			pkgerr.FieldError{Field: "currency_code", Issue: "required"},
		)
	}
	r, err := h.regionService.CreateRegion(c.Context(), region.RegionInput{
		Name: req.Name, Locale: req.Locale, CurrencyCode: req.CurrencyCode,
		TaxInclusive: req.TaxInclusive, Countries: req.Countries,
	})
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"region": toRegionResponse(r)})
}

func (h *RegionHandler) UpdateRegion(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	var req dto.CreateRegionRequest
	if err := c.Bind().Body(&req); err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	r, err := h.regionService.UpdateRegion(c.Context(), id, region.RegionInput{
		Name: req.Name, Locale: req.Locale, CurrencyCode: req.CurrencyCode,
		TaxInclusive: req.TaxInclusive, Countries: req.Countries,
	})
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"region": toRegionResponse(r)})
}

func (h *RegionHandler) DeleteRegion(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	if err := h.regionService.DeleteRegion(c.Context(), id); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *RegionHandler) ListTaxRates(c fiber.Ctx) error {
	regionID, _ := strconv.Atoi(c.Query("region_id", "0"))
	rates, err := h.regionService.ListTaxRates(c.Context(), regionID)
	if err != nil {
		return err
	}
	var resp []dto.TaxRateResponse
	for _, t := range rates {
		resp = append(resp, dto.TaxRateResponse{
			ID: t.ID, RegionID: t.RegionID, CountryCode: t.CountryCode,
			Rate: t.Rate, Name: t.Name, Priority: t.Priority,
		})
	}
	return c.JSON(fiber.Map{"data": resp})
}

func (h *RegionHandler) CreateTaxRate(c fiber.Ctx) error {
	var req dto.CreateTaxRateRequest
	if err := c.Bind().Body(&req); err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	if req.RegionID == 0 || req.Name == "" {
		return pkgerr.ErrValidation.WithDetails(
			pkgerr.FieldError{Field: "region_id", Issue: "required"},
			pkgerr.FieldError{Field: "name", Issue: "required"},
		)
	}
	t, err := h.regionService.CreateTaxRate(c.Context(), region.TaxRateInput{
		RegionID: req.RegionID, CountryCode: req.CountryCode,
		Rate: req.Rate, Name: req.Name, Priority: req.Priority,
	})
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"tax_rate": dto.TaxRateResponse{
		ID: t.ID, RegionID: t.RegionID, CountryCode: t.CountryCode,
		Rate: t.Rate, Name: t.Name, Priority: t.Priority,
	}})
}

func (h *RegionHandler) DeleteTaxRate(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	if err := h.regionService.DeleteTaxRate(c.Context(), id); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func toRegionResponse(r *ent.Region) dto.RegionResponse {
	return dto.RegionResponse{
		ID: r.ID, Name: r.Name, Locale: r.Locale, CurrencyCode: r.CurrencyCode,
		TaxInclusive: r.TaxInclusive, Countries: r.Countries, Status: string(r.Status),
	}
}
