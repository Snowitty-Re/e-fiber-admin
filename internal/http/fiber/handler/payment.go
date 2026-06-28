package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"

	paysvc "github.com/Snowitty-Re/e-fiber-admin/internal/domain/payment"
	"github.com/Snowitty-Re/e-fiber-admin/internal/ent"
	"github.com/Snowitty-Re/e-fiber-admin/internal/http/fiber/dto"
	pkgerr "github.com/Snowitty-Re/e-fiber-admin/internal/pkg/errors"
)

type PaymentHandler struct {
	paymentService *paysvc.Service
}

func NewPaymentHandler(ps *paysvc.Service) *PaymentHandler {
	return &PaymentHandler{paymentService: ps}
}

func (h *PaymentHandler) AdminListProviders(c fiber.Ctx) error {
	providers, err := h.paymentService.ListProviders(c.Context())
	if err != nil {
		return err
	}
	var data []dto.PaymentProviderResponse
	for _, p := range providers {
		data = append(data, dto.PaymentProviderResponse{
			ID: p.ID, Code: p.Code, Name: p.Name, IsActive: p.IsActive,
		})
	}
	return c.JSON(fiber.Map{"data": data})
}

func (h *PaymentHandler) AdminCreateProvider(c fiber.Ctx) error {
	var req dto.CreatePaymentProviderRequest
	if err := c.Bind().Body(&req); err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	if req.Code == "" || req.Name == "" {
		return pkgerr.ErrValidation.WithDetails(
			pkgerr.FieldError{Field: "code", Issue: "required"},
			pkgerr.FieldError{Field: "name", Issue: "required"},
		)
	}
	p, err := h.paymentService.CreateProvider(c.Context(), req.Code, req.Name, req.Config)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"provider": dto.PaymentProviderResponse{
		ID: p.ID, Code: p.Code, Name: p.Name, IsActive: p.IsActive,
	}})
}

func (h *PaymentHandler) AdminCreateSession(c fiber.Ctx) error {
	var req dto.CreatePaymentSessionRequest
	if err := c.Bind().Body(&req); err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	session, err := h.paymentService.CreateSession(c.Context(), req.OrderID, req.ProviderCode, "USD", 0)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"session": toSessionResponse(session)})
}

func (h *PaymentHandler) AdminAuthorize(c fiber.Ctx) error {
	var req dto.AuthorizeRequest
	if err := c.Bind().Body(&req); err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	session, err := h.paymentService.Authorize(c.Context(), req.SessionID)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"session": toSessionResponse(session)})
}

func (h *PaymentHandler) AdminCapture(c fiber.Ctx) error {
	var req dto.CaptureRequest
	if err := c.Bind().Body(&req); err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	if err := h.paymentService.Capture(c.Context(), req.SessionID, req.Amount); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *PaymentHandler) AdminRefund(c fiber.Ctx) error {
	sessionID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	amount, _ := strconv.ParseInt(c.Query("amount", "0"), 10, 64)
	if err := h.paymentService.Refund(c.Context(), sessionID, amount); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *PaymentHandler) AdminGetSession(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	session, err := h.paymentService.GetSession(c.Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"session": toSessionResponse(session)})
}

func (h *PaymentHandler) StoreCreateSession(c fiber.Ctx) error {
	var req dto.CreatePaymentSessionRequest
	if err := c.Bind().Body(&req); err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	amountStr := c.Query("amount", "0")
	amount, _ := strconv.ParseInt(amountStr, 10, 64)
	currency := c.Get("X-Currency", "USD")
	session, err := h.paymentService.CreateSession(c.Context(), req.OrderID, req.ProviderCode, currency, amount)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"session": toSessionResponse(session)})
}

func toSessionResponse(s *ent.PaymentSession) dto.PaymentSessionResponse {
	return dto.PaymentSessionResponse{
		ID: s.ID, OrderID: s.OrderID, ProviderCode: s.ProviderCode,
		Status: string(s.Status), Amount: s.Amount, CurrencyCode: s.CurrencyCode,
	}
}
