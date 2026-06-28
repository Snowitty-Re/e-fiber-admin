package handler

import (
	"github.com/gofiber/fiber/v3"

	shipsvc "github.com/Snowitty-Re/e-fiber-admin/internal/domain/shipping"
	"github.com/Snowitty-Re/e-fiber-admin/internal/ent"
	"github.com/Snowitty-Re/e-fiber-admin/internal/http/fiber/dto"
	pkgerr "github.com/Snowitty-Re/e-fiber-admin/internal/pkg/errors"
)

type ShippingHandler struct {
	shippingService *shipsvc.Service
}

func NewShippingHandler(ss *shipsvc.Service) *ShippingHandler {
	return &ShippingHandler{shippingService: ss}
}

func (h *ShippingHandler) AdminListProfiles(c fiber.Ctx) error {
	profiles, err := h.shippingService.ListProfiles(c.Context())
	if err != nil {
		return err
	}
	var data []dto.ShippingProfileResponse
	for _, p := range profiles {
		resp := dto.ShippingProfileResponse{
			ID: p.ID, Name: p.Name, ProductType: p.ProductType,
			Options: []dto.ShippingOptionResponse{},
		}
		for _, o := range p.Edges.Options {
			resp.Options = append(resp.Options, dto.ShippingOptionResponse{
				ID: o.ID, ProfileID: o.ProfileID, Name: o.Name,
				PriceAmount: o.PriceAmount, PriceCurrency: o.PriceCurrency,
				EstimatedDays: o.EstimatedDays, IsActive: o.IsActive,
			})
		}
		data = append(data, resp)
	}
	return c.JSON(fiber.Map{"data": data})
}

func (h *ShippingHandler) AdminCreateProfile(c fiber.Ctx) error {
	var req dto.CreateShippingProfileRequest
	if err := c.Bind().Body(&req); err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	if req.Name == "" {
		return pkgerr.ErrValidation.WithDetails(pkgerr.FieldError{Field: "name", Issue: "required"})
	}
	pt := req.ProductType
	if pt == "" {
		pt = "physical"
	}
	p, err := h.shippingService.CreateProfile(c.Context(), req.Name, pt)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"profile": dto.ShippingProfileResponse{
		ID: p.ID, Name: p.Name, ProductType: p.ProductType, Options: []dto.ShippingOptionResponse{},
	}})
}

func (h *ShippingHandler) AdminCreateOption(c fiber.Ctx) error {
	var req dto.CreateShippingOptionRequest
	if err := c.Bind().Body(&req); err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	currency := req.PriceCurrency
	if currency == "" {
		currency = "USD"
	}
	o, err := h.shippingService.CreateOption(c.Context(), req.ProfileID, req.Name, currency, req.PriceAmount, req.EstimatedDays, req.IsActive)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"option": dto.ShippingOptionResponse{
		ID: o.ID, ProfileID: o.ProfileID, Name: o.Name,
		PriceAmount: o.PriceAmount, PriceCurrency: o.PriceCurrency,
		EstimatedDays: o.EstimatedDays, IsActive: o.IsActive,
	}})
}

func (h *ShippingHandler) StoreQuote(c fiber.Ctx) error {
	currency := c.Get("X-Currency", "USD")
	options, err := h.shippingService.Quote(c.Context(), currency)
	if err != nil {
		return err
	}
	var data []dto.ShippingOptionResponse
	for _, o := range options {
		data = append(data, dto.ShippingOptionResponse{
			ID: o.ID, ProfileID: o.ProfileID, Name: o.Name,
			PriceAmount: o.PriceAmount, PriceCurrency: o.PriceCurrency,
			EstimatedDays: o.EstimatedDays, IsActive: o.IsActive,
		})
	}
	return c.JSON(fiber.Map{"data": data})
}

var _ = ent.ShippingProfile{}
