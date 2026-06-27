package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"

	ordersvc "github.com/Snowitty-Re/e-fiber-admin/internal/domain/order"
	"github.com/Snowitty-Re/e-fiber-admin/internal/ent"
	"github.com/Snowitty-Re/e-fiber-admin/internal/http/fiber/dto"
	pkgerr "github.com/Snowitty-Re/e-fiber-admin/internal/pkg/errors"
)

type OrderHandler struct {
	orderService *ordersvc.Service
	entClient    *ent.Client
}

func NewOrderHandler(os *ordersvc.Service, entClient *ent.Client) *OrderHandler {
	return &OrderHandler{orderService: os, entClient: entClient}
}

func (h *OrderHandler) StoreCheckout(c fiber.Ctx) error {
	var req dto.CheckoutRequest
	if err := c.Bind().Body(&req); err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	if req.CartID == 0 || req.Email == "" {
		return pkgerr.ErrValidation.WithDetails(
			pkgerr.FieldError{Field: "cart_id", Issue: "required"},
			pkgerr.FieldError{Field: "email", Issue: "required"},
		)
	}
	store, err := h.entClient.Store.Query().First(c.Context())
	if err != nil {
		return pkgerr.ErrServiceUnavailable.WithCause(err)
	}
	customerID, _ := c.Locals("customer_id").(int64)
	result, err := h.orderService.CompleteCheckout(c.Context(), ordersvc.CheckoutInput{
		CartID: req.CartID, Email: req.Email, CustomerID: int(customerID),
		CurrencyCode: store.DefaultCurrency, Locale: store.DefaultLocale,
		ShippingAddress: req.ShippingAddress, BillingAddress: req.BillingAddress,
	})
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"order": toOrderResponse(result.Order)})
}

func (h *OrderHandler) StoreListOrders(c fiber.Ctx) error {
	customerID, ok := c.Locals("customer_id").(int64)
	if !ok {
		return pkgerr.ErrUnauthorized
	}
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))
	items, total, err := h.orderService.List(c.Context(), page, pageSize, "", int(customerID))
	if err != nil {
		return err
	}
	var data []dto.OrderResponse
	for _, o := range items {
		data = append(data, toOrderResponse(o))
	}
	return c.JSON(fiber.Map{
		"data": data,
		"pagination": fiber.Map{"page": page, "page_size": pageSize, "total": total,
			"has_more": int64(page*pageSize) < int64(total)},
	})
}

func (h *OrderHandler) StoreGetOrder(c fiber.Ctx) error {
	customerID, ok := c.Locals("customer_id").(int64)
	if !ok {
		return pkgerr.ErrUnauthorized
	}
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	o, err := h.orderService.Get(c.Context(), id)
	if err != nil {
		return err
	}
	if int64(o.CustomerID) != customerID {
		return pkgerr.ErrNotFound
	}
	return c.JSON(fiber.Map{"order": toOrderResponse(o)})
}

func (h *OrderHandler) AdminList(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))
	status := c.Query("status", "")
	items, total, err := h.orderService.List(c.Context(), page, pageSize, status, 0)
	if err != nil {
		return err
	}
	var data []dto.OrderResponse
	for _, o := range items {
		data = append(data, toOrderResponse(o))
	}
	return c.JSON(fiber.Map{
		"data": data,
		"pagination": fiber.Map{"page": page, "page_size": pageSize, "total": total,
			"has_more": int64(page*pageSize) < int64(total)},
	})
}

func (h *OrderHandler) AdminGet(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	o, err := h.orderService.Get(c.Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"order": toOrderResponse(o)})
}

func (h *OrderHandler) AdminCancel(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	if err := h.orderService.Cancel(c.Context(), id); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func toOrderResponse(o *ent.Order) dto.OrderResponse {
	resp := dto.OrderResponse{
		ID: o.ID, Number: o.Number, Email: o.Email, CustomerID: o.CustomerID,
		CurrencyCode: o.CurrencyCode, Status: string(o.Status),
		FulfillmentStatus: string(o.FulfillmentStatus), PaymentStatus: string(o.PaymentStatus),
		Totals: o.Totals, ShippingAddress: o.ShippingAddress,
		Items: []dto.OrderItemResponse{},
	}
	if o.PlacedAt != nil {
		resp.PlacedAt = &[]string{o.PlacedAt.Format("2006-01-02T15:04:05Z")}[0]
	}
	for _, item := range o.Edges.Items {
		vID := 0
		if item.VariantID > 0 {
			vID = item.VariantID
		}
		resp.Items = append(resp.Items, dto.OrderItemResponse{
			ID: item.ID, VariantID: vID, SKU: item.Sku, Title: item.Title,
			Quantity: item.Quantity, UnitAmount: item.UnitAmount,
			TotalAmount: item.TotalAmount,
		})
	}
	return resp
}

func (h *OrderHandler) AdminCreateFulfillment(c fiber.Ctx) error {
	orderID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	var req dto.CreateFulfillmentRequest
	if err := c.Bind().Body(&req); err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	in := ordersvc.FulfillmentInput{
		OrderID: orderID, TrackingNumber: req.TrackingNumber,
	}
	for _, item := range req.Items {
		in.Items = append(in.Items, ordersvc.FulfillmentItemInput{
			OrderItemID: item.OrderItemID, Quantity: item.Quantity,
		})
	}
	f, err := h.orderService.CreateFulfillment(c.Context(), in)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"fulfillment": fiber.Map{
		"id": f.ID, "order_id": f.OrderID, "tracking_number": f.TrackingNumber,
		"status": string(f.Status), "items": len(f.Edges.Items),
	}})
}

func (h *OrderHandler) AdminListFulfillments(c fiber.Ctx) error {
	orderID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	fulfillments, err := h.orderService.ListFulfillments(c.Context(), orderID)
	if err != nil {
		return err
	}
	var data []map[string]any
	for _, f := range fulfillments {
		data = append(data, map[string]any{
			"id": f.ID, "order_id": f.OrderID, "tracking_number": f.TrackingNumber,
			"status": string(f.Status), "items": len(f.Edges.Items),
		})
	}
	return c.JSON(fiber.Map{"data": data})
}

func (h *OrderHandler) AdminMarkPaid(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	if err := h.orderService.MarkPaid(c.Context(), id); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *OrderHandler) AdminCreateReturn(c fiber.Ctx) error {
	orderID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	var req dto.CreateReturnRequest
	if err := c.Bind().Body(&req); err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	in := ordersvc.ReturnInput{
		OrderID: orderID, Reason: req.Reason,
		RefundAmount: req.RefundAmount, CurrencyCode: req.CurrencyCode,
	}
	for _, item := range req.Items {
		in.Items = append(in.Items, ordersvc.ReturnItemInput{
			OrderItemID: item.OrderItemID, Quantity: item.Quantity, Reason: item.Reason,
		})
	}
	r, err := h.orderService.CreateReturn(c.Context(), in)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"return": fiber.Map{
		"id": r.ID, "order_id": r.OrderID, "status": string(r.Status),
		"refund_amount": r.RefundAmount, "items": len(r.Edges.Items),
	}})
}
