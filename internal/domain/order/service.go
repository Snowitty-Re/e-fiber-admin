package order

import (
	"context"
	"fmt"
	"time"

	"github.com/Snowitty-Re/e-fiber-admin/internal/ent"
	"github.com/Snowitty-Re/e-fiber-admin/internal/ent/cart"
	"github.com/Snowitty-Re/e-fiber-admin/internal/ent/order"
	"github.com/Snowitty-Re/e-fiber-admin/internal/ent/variant"
	"github.com/Snowitty-Re/e-fiber-admin/internal/events"
	pkgerr "github.com/Snowitty-Re/e-fiber-admin/internal/pkg/errors"
)

type Service struct {
	entClient *ent.Client
	bus       *events.Bus
}

func NewService(entClient *ent.Client, bus *events.Bus) *Service {
	return &Service{entClient: entClient, bus: bus}
}

type CheckoutInput struct {
	CartID          int
	Email           string
	CustomerID      int
	CurrencyCode    string
	Locale          string
	ShippingAddress map[string]any
	BillingAddress  map[string]any
}

type OrderResult struct {
	Order     *ent.Order
	OrderID   int
	OrderNumber string
}

func (s *Service) CompleteCheckout(ctx context.Context, in CheckoutInput) (*OrderResult, error) {
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	c, err := tx.Cart.Query().
		Where(cart.IDEQ(in.CartID), cart.StatusEQ(cart.StatusActive)).
		WithItems().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, pkgerr.New("CART_NOT_FOUND", 404, "cart not found or not active")
		}
		return nil, fmt.Errorf("query cart: %w", err)
	}
	if len(c.Edges.Items) == 0 {
		return nil, pkgerr.New("CART_EMPTY", 409, "cart is empty")
	}

	var total int64
	for _, item := range c.Edges.Items {
		total += item.UnitAmount * int64(item.Quantity)
		v, err := tx.Variant.Query().Where(variant.IDEQ(item.VariantID)).Only(ctx)
		if err != nil {
			return nil, fmt.Errorf("query variant %d: %w", item.VariantID, err)
		}
		if v.Inventory < item.Quantity && !v.AllowBackorder {
			return nil, pkgerr.New("VARIANT_OUT_OF_STOCK", 409,
				fmt.Sprintf("variant %s (id=%d) has only %d in stock", v.Sku, v.ID, v.Inventory))
		}
		_, err = v.Update().SetInventory(v.Inventory - item.Quantity).Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("deduct inventory for variant %d: %w", item.VariantID, err)
		}
	}

	orderNumber := generateOrderNumber()
	o, err := tx.Order.Create().
		SetNumber(orderNumber).
		SetEmail(in.Email).
		SetCurrencyCode(in.CurrencyCode).
		SetLocale(in.Locale).
		SetStatus(order.StatusPending).
		SetFulfillmentStatus(order.FulfillmentStatusNotFulfilled).
		SetPaymentStatus(order.PaymentStatusAwaiting).
		SetShippingAddress(in.ShippingAddress).
		SetBillingAddress(in.BillingAddress).
		SetTotals(map[string]any{"subtotal": total, "total": total, "currency": in.CurrencyCode}).
		SetPlacedAt(time.Now()).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}
	if in.CustomerID > 0 {
		_, err = o.Update().SetCustomerID(in.CustomerID).Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("set customer: %w", err)
		}
	}

	for _, item := range c.Edges.Items {
		lineTotal := item.UnitAmount * int64(item.Quantity)
		_, err = tx.OrderItem.Create().
			SetOrderID(o.ID).
			SetVariantID(item.VariantID).
			SetSku(item.Sku).
			SetTitle(item.Sku).
			SetQuantity(item.Quantity).
			SetUnitAmount(item.UnitAmount).
			SetTotalAmount(lineTotal).
			Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("create order item: %w", err)
		}
	}

	_, err = tx.Cart.UpdateOneID(c.ID).SetStatus(cart.StatusConverted).Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("mark cart converted: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	fullOrder, err := s.Get(ctx, o.ID)
	if err != nil {
		return nil, err
	}

	if s.bus != nil {
		_ = s.bus.PublishSimple(ctx, "order.placed", "order", fmt.Sprintf("%d", o.ID), map[string]any{
			"order_id": o.ID, "order_number": orderNumber,
			"email": in.Email, "total": total, "currency": in.CurrencyCode,
		})
	}

	return &OrderResult{Order: fullOrder, OrderID: o.ID, OrderNumber: orderNumber}, nil
}

func (s *Service) Get(ctx context.Context, id int) (*ent.Order, error) {
	o, err := s.entClient.Order.Query().
		Where(order.IDEQ(id)).
		WithItems().
		WithFulfillments(func(q *ent.FulfillmentQuery) { q.WithItems() }).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, pkgerr.ErrNotFound
		}
		return nil, fmt.Errorf("query order: %w", err)
	}
	return o, nil
}

func (s *Service) GetByNumber(ctx context.Context, number string) (*ent.Order, error) {
	o, err := s.entClient.Order.Query().
		Where(order.NumberEQ(number)).
		WithItems().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, pkgerr.ErrNotFound
		}
		return nil, fmt.Errorf("query order: %w", err)
	}
	return o, nil
}

func (s *Service) List(ctx context.Context, page, pageSize int, status string, customerID int) ([]*ent.Order, int, error) {
	q := s.entClient.Order.Query()
	if status != "" {
		q = q.Where(order.StatusEQ(order.Status(status)))
	}
	if customerID > 0 {
		q = q.Where(order.CustomerIDEQ(customerID))
	}
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	items, err := q.WithItems().
		Order(ent.Desc(order.FieldID)).
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		All(ctx)
	return items, total, err
}

func (s *Service) Cancel(ctx context.Context, id int) error {
	o, err := s.Get(ctx, id)
	if err != nil {
		return err
	}
	if o.Status != order.StatusPending && o.Status != order.StatusPaid {
		return pkgerr.New("ORDER_INVALID_STATE", 409, fmt.Sprintf("cannot cancel order in status %s", o.Status))
	}

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	for _, item := range o.Edges.Items {
		if item.VariantID > 0 {
			v, err := tx.Variant.Query().Where(variant.IDEQ(item.VariantID)).Only(ctx)
			if err == nil {
				_, _ = v.Update().SetInventory(v.Inventory + item.Quantity).Save(ctx)
			}
		}
	}

	err = tx.Order.UpdateOneID(id).
		SetStatus(order.StatusCancelled).
		SetCancelledAt(time.Now()).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("cancel order: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	if s.bus != nil {
		_ = s.bus.PublishSimple(ctx, "order.cancelled", "order", fmt.Sprintf("%d", id), map[string]any{
			"order_id": id,
		})
	}
	return nil
}

func (s *Service) MarkPaid(ctx context.Context, id int) error {
	if err := s.entClient.Order.UpdateOneID(id).
		SetPaymentStatus(order.PaymentStatusPaid).
		SetStatus(order.StatusPaid).
		Exec(ctx); err != nil {
		return err
	}
	if s.bus != nil {
		_ = s.bus.PublishSimple(ctx, "order.paid", "order", fmt.Sprintf("%d", id), map[string]any{
			"order_id": id,
		})
	}
	return nil
}

func generateOrderNumber() string {
	return fmt.Sprintf("EFA-%s-%d", time.Now().UTC().Format("20060102"), time.Now().UnixNano()%10000)
}