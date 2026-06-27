package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"

	cartsvc "github.com/Snowitty-Re/e-fiber-admin/internal/domain/cart"
	"github.com/Snowitty-Re/e-fiber-admin/internal/ent"
	"github.com/Snowitty-Re/e-fiber-admin/internal/http/fiber/dto"
	pkgerr "github.com/Snowitty-Re/e-fiber-admin/internal/pkg/errors"
)

type CartHandler struct {
	cartService *cartsvc.Service
	entClient   *ent.Client
}

func NewCartHandler(cs *cartsvc.Service, entClient *ent.Client) *CartHandler {
	return &CartHandler{cartService: cs, entClient: entClient}
}

func (h *CartHandler) CreateCart(c fiber.Ctx) error {
	store, err := h.entClient.Store.Query().First(c.Context())
	if err != nil {
		return pkgerr.ErrServiceUnavailable.WithCause(err)
	}
	currency := c.Get("X-Currency", store.DefaultCurrency)
	if currency == "" {
		currency = store.DefaultCurrency
	}
	locale := store.DefaultLocale

	var req dto.CreateCartRequest
	_ = c.Bind().Body(&req)
	if req.CurrencyCode != "" {
		currency = req.CurrencyCode
	}
	if req.Locale != "" {
		locale = req.Locale
	}

	customerID, _ := c.Locals("customer_id").(int64)
	cart, err := h.cartService.CreateCart(c.Context(), int(customerID), "", currency, locale)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"cart": toCartResponse(cart)})
}

func (h *CartHandler) GetCart(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	cart, err := h.cartService.GetCart(c.Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"cart": toCartResponse(cart)})
}

func (h *CartHandler) AddItem(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	var req dto.AddCartItemRequest
	if err := c.Bind().Body(&req); err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	if req.VariantID == 0 || req.Quantity < 1 {
		return pkgerr.ErrValidation.WithDetails(
			pkgerr.FieldError{Field: "variant_id", Issue: "required"},
			pkgerr.FieldError{Field: "quantity", Issue: "must be >= 1"},
		)
	}
	store, err := h.entClient.Store.Query().First(c.Context())
	if err != nil {
		return pkgerr.ErrServiceUnavailable.WithCause(err)
	}
	currency := c.Get("X-Currency", store.DefaultCurrency)
	cart, err := h.cartService.AddItem(c.Context(), cartsvc.AddItemInput{
		CartID: id, VariantID: req.VariantID, Quantity: req.Quantity,
		CurrencyCode: currency,
	})
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"cart": toCartResponse(cart)})
}

func (h *CartHandler) UpdateItem(c fiber.Ctx) error {
	cartID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	variantID, err := strconv.Atoi(c.Params("variantId"))
	if err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	var req dto.UpdateCartItemRequest
	if err := c.Bind().Body(&req); err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	cart, err := h.cartService.UpdateItemQuantity(c.Context(), cartID, variantID, req.Quantity)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"cart": toCartResponse(cart)})
}

func (h *CartHandler) RemoveItem(c fiber.Ctx) error {
	cartID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	variantID, err := strconv.Atoi(c.Params("variantId"))
	if err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	cart, err := h.cartService.RemoveItem(c.Context(), cartID, variantID)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"cart": toCartResponse(cart)})
}

func toCartResponse(cart *ent.Cart) dto.CartResponse {
	resp := dto.CartResponse{
		ID: cart.ID, Status: string(cart.Status),
		Currency: cart.CurrencyCode, Locale: cart.Locale,
		Items: []dto.CartItemResponse{},
	}
	var total int64
	for _, item := range cart.Edges.Items {
		lineTotal := item.UnitAmount * int64(item.Quantity)
		resp.Items = append(resp.Items, dto.CartItemResponse{
			ID: item.ID, VariantID: item.VariantID, ProductID: item.ProductID,
			SKU: item.Sku, Quantity: item.Quantity,
			UnitAmount: item.UnitAmount, TotalAmount: lineTotal,
			Currency: item.CurrencyCode,
		})
		total += lineTotal
	}
	resp.TotalAmount = total
	return resp
}