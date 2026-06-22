package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"

	customersvc "github.com/Snowitty-Re/e-fiber-admin/internal/domain/customer"
	"github.com/Snowitty-Re/e-fiber-admin/internal/ent"
	"github.com/Snowitty-Re/e-fiber-admin/internal/http/fiber/dto"
	pkgerr "github.com/Snowitty-Re/e-fiber-admin/internal/pkg/errors"
)

type CustomerHandler struct {
	customerService *customersvc.Service
}

func NewCustomerHandler(cs *customersvc.Service) *CustomerHandler {
	return &CustomerHandler{customerService: cs}
}

func (h *CustomerHandler) Register(c fiber.Ctx) error {
	var req dto.CustomerRegisterRequest
	if err := c.Bind().Body(&req); err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	if req.Email == "" || req.Password == "" {
		return pkgerr.ErrValidation.WithDetails(
			pkgerr.FieldError{Field: "email", Issue: "required"},
			pkgerr.FieldError{Field: "password", Issue: "required"},
		)
	}
	identity, pair, err := h.customerService.Register(c.Context(), customersvc.RegisterInput{
		Email: req.Email, Password: req.Password, FirstName: req.FirstName,
		LastName: req.LastName, DefaultCurrency: req.DefaultCurrency,
		DefaultLocale: req.DefaultLocale,
	})
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(dto.CustomerTokenResponse{
		AccessToken: pair.AccessToken, RefreshToken: pair.RefreshToken,
		ExpiresIn: pair.ExpiresIn,
		Customer: dto.CustomerProfile{
			ID: int(identity.ID), Email: identity.Email,
			FirstName: identity.FirstName, LastName: identity.LastName,
			DefaultCurrency: identity.DefaultCurrency, DefaultLocale: identity.DefaultLocale,
		},
	})
}

func (h *CustomerHandler) Login(c fiber.Ctx) error {
	var req dto.CustomerLoginRequest
	if err := c.Bind().Body(&req); err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	identity, pair, err := h.customerService.Login(c.Context(), req.Email, req.Password)
	if err != nil {
		return err
	}
	return c.JSON(dto.CustomerTokenResponse{
		AccessToken: pair.AccessToken, RefreshToken: pair.RefreshToken,
		ExpiresIn: pair.ExpiresIn,
		Customer: dto.CustomerProfile{
			ID: int(identity.ID), Email: identity.Email,
			FirstName: identity.FirstName, LastName: identity.LastName,
			DefaultCurrency: identity.DefaultCurrency, DefaultLocale: identity.DefaultLocale,
		},
	})
}

func (h *CustomerHandler) Refresh(c fiber.Ctx) error {
	var req dto.CustomerRefreshRequest
	if err := c.Bind().Body(&req); err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	pair, err := h.customerService.Refresh(c.Context(), req.RefreshToken)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{
		"access_token": pair.AccessToken, "refresh_token": pair.RefreshToken,
		"expires_in": pair.ExpiresIn,
	})
}

func (h *CustomerHandler) Me(c fiber.Ctx) error {
	customerID, ok := c.Locals("customer_id").(int64)
	if !ok {
		return pkgerr.ErrUnauthorized
	}
	identity, err := h.customerService.Me(c.Context(), customerID)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"customer": dto.CustomerProfile{
		ID: int(identity.ID), Email: identity.Email,
		FirstName: identity.FirstName, LastName: identity.LastName,
		DefaultCurrency: identity.DefaultCurrency, DefaultLocale: identity.DefaultLocale,
	}})
}

func (h *CustomerHandler) Logout(c fiber.Ctx) error {
	var req dto.CustomerRefreshRequest
	_ = c.Bind().Body(&req)
	_ = h.customerService.Logout(c.Context(), req.RefreshToken)
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *CustomerHandler) AdminList(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))
	search := c.Query("q", "")
	items, total, err := h.customerService.List(c.Context(), page, pageSize, search)
	if err != nil {
		return err
	}
	var data []dto.CustomerAdminResponse
	for _, c := range items {
		data = append(data, dto.CustomerAdminResponse{
			ID: c.ID, Email: c.Email, Status: string(c.Status),
		})
	}
	return c.JSON(fiber.Map{
		"data": data,
		"pagination": fiber.Map{"page": page, "page_size": pageSize, "total": total,
			"has_more": int64(page*pageSize) < int64(total)},
	})
}

func (h *CustomerHandler) AdminGet(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	cust, err := h.customerService.Get(c.Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"customer": fiber.Map{
		"id": cust.ID, "email": cust.Email, "status": string(cust.Status),
		"first_name": cust.FirstName, "last_name": cust.LastName,
		"phone": cust.Phone,
		"default_currency": cust.DefaultCurrency, "default_locale": cust.DefaultLocale,
	}})
}

func (h *CustomerHandler) AdminDisable(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	if err := h.customerService.Disable(c.Context(), id); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

var _ = ent.Customer{}