package cart

import (
	"context"
	"fmt"
	"time"

	"github.com/Snowitty-Re/e-fiber-admin/internal/ent"
	"github.com/Snowitty-Re/e-fiber-admin/internal/ent/cart"
	"github.com/Snowitty-Re/e-fiber-admin/internal/ent/cartitem"
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

type AddItemInput struct {
	CartID       int
	VariantID    int
	Quantity     int
	CurrencyCode string
}

func (s *Service) CreateCart(ctx context.Context, customerID int, email, currencyCode, locale string) (*ent.Cart, error) {
	b := s.entClient.Cart.Create().
		SetCurrencyCode(currencyCode).
		SetLocale(locale).
		SetStatus(cart.StatusActive)
	if customerID > 0 {
		b.SetCustomerID(customerID)
	}
	if email != "" {
		b.SetEmail(email)
	}
	return b.Save(ctx)
}

func (s *Service) GetCart(ctx context.Context, id int) (*ent.Cart, error) {
	c, err := s.entClient.Cart.Query().
		Where(cart.IDEQ(id)).
		WithItems().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, pkgerr.ErrNotFound
		}
		return nil, fmt.Errorf("query cart: %w", err)
	}
	return c, nil
}

func (s *Service) AddItem(ctx context.Context, in AddItemInput) (*ent.Cart, error) {
	v, err := s.entClient.Variant.Query().
		Where(variant.IDEQ(in.VariantID)).
		WithPrices().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, pkgerr.New("VARIANT_NOT_FOUND", 404, "variant not found")
		}
		return nil, fmt.Errorf("query variant: %w", err)
	}

	var price int64
	for _, p := range v.Edges.Prices {
		if p.CurrencyCode == in.CurrencyCode {
			price = p.Amount
			break
		}
	}
	if price == 0 {
		return nil, pkgerr.New("VARIANT_PRICE_NOT_FOUND", 409, "no price for currency "+in.CurrencyCode)
	}

	if v.Inventory < in.Quantity && !v.AllowBackorder {
		return nil, pkgerr.New("VARIANT_OUT_OF_STOCK", 409, fmt.Sprintf("only %d in stock", v.Inventory))
	}

	existing, err := s.entClient.CartItem.Query().
		Where(cartitem.CartIDEQ(in.CartID), cartitem.VariantIDEQ(in.VariantID)).
		Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return nil, fmt.Errorf("query existing item: %w", err)
	}
	if existing != nil {
		_, err = existing.Update().
			SetQuantity(existing.Quantity + in.Quantity).
			SetUnitAmount(price).
			Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("update cart item: %w", err)
		}
	} else {
		_, err = s.entClient.CartItem.Create().
			SetCartID(in.CartID).
			SetVariantID(in.VariantID).
			SetProductID(v.ProductID).
			SetQuantity(in.Quantity).
			SetUnitAmount(price).
			SetCurrencyCode(in.CurrencyCode).
			SetSku(v.Sku).
			Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("create cart item: %w", err)
		}
	}

	return s.GetCart(ctx, in.CartID)
}

func (s *Service) UpdateItemQuantity(ctx context.Context, cartID, variantID, quantity int) (*ent.Cart, error) {
	if quantity <= 0 {
		return s.RemoveItem(ctx, cartID, variantID)
	}
	_, err := s.entClient.CartItem.Update().
		Where(cartitem.CartIDEQ(cartID), cartitem.VariantIDEQ(variantID)).
		SetQuantity(quantity).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("update quantity: %w", err)
	}
	return s.GetCart(ctx, cartID)
}

func (s *Service) RemoveItem(ctx context.Context, cartID, variantID int) (*ent.Cart, error) {
	_, err := s.entClient.CartItem.Delete().
		Where(cartitem.CartIDEQ(cartID), cartitem.VariantIDEQ(variantID)).
		Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("remove item: %w", err)
	}
	return s.GetCart(ctx, cartID)
}

func (s *Service) CartTotal(ctx context.Context, id int) (int64, error) {
	c, err := s.GetCart(ctx, id)
	if err != nil {
		return 0, err
	}
	var total int64
	for _, item := range c.Edges.Items {
		total += item.UnitAmount * int64(item.Quantity)
	}
	return total, nil
}

func (s *Service) MarkConverted(ctx context.Context, id int) error {
	return s.entClient.Cart.UpdateOneID(id).
		SetStatus(cart.StatusConverted).
		Exec(ctx)
}

func (s *Service) MarkAbandoned(ctx context.Context, id int) error {
	return s.entClient.Cart.UpdateOneID(id).
		SetStatus(cart.StatusAbandoned).
		Exec(ctx)
}

func (s *Service) FindAbandonedCarts(ctx context.Context, before time.Time) ([]*ent.Cart, error) {
	return s.entClient.Cart.Query().
		Where(cart.StatusEQ(cart.StatusActive), cart.UpdatedAtLT(before)).
		All(ctx)
}
