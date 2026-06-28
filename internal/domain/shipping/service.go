package shipping

import (
	"context"
	"fmt"

	"github.com/Snowitty-Re/e-fiber-admin/internal/ent"
	"github.com/Snowitty-Re/e-fiber-admin/internal/ent/shippingoption"
	"github.com/Snowitty-Re/e-fiber-admin/internal/events"
)

type Service struct {
	entClient *ent.Client
	bus       *events.Bus
}

func NewService(entClient *ent.Client, bus *events.Bus) *Service {
	return &Service{entClient: entClient, bus: bus}
}

func (s *Service) ListProfiles(ctx context.Context) ([]*ent.ShippingProfile, error) {
	return s.entClient.ShippingProfile.Query().WithOptions().All(ctx)
}

func (s *Service) CreateProfile(ctx context.Context, name, productType string) (*ent.ShippingProfile, error) {
	return s.entClient.ShippingProfile.Create().
		SetName(name).
		SetProductType(productType).
		Save(ctx)
}

func (s *Service) CreateOption(ctx context.Context, profileID int, name, currency string, priceAmount, estimatedDays int64, isActive bool) (*ent.ShippingOption, error) {
	return s.entClient.ShippingOption.Create().
		SetProfileID(profileID).
		SetName(name).
		SetPriceAmount(priceAmount).
		SetPriceCurrency(currency).
		SetEstimatedDays(int(estimatedDays)).
		SetIsActive(isActive).
		Save(ctx)
}

func (s *Service) Quote(ctx context.Context, currencyCode string) ([]*ent.ShippingOption, error) {
	options, err := s.entClient.ShippingOption.Query().
		Where(shippingoption.IsActiveEQ(true), shippingoption.PriceCurrencyEQ(currencyCode)).
		WithProfile().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query shipping options: %w", err)
	}
	if s.bus != nil {
		_ = s.bus.PublishSimple(ctx, "shipping_option.quoted", "shipping", "", map[string]any{
			"count": len(options), "currency": currencyCode,
		})
	}
	return options, nil
}

func (s *Service) AssignProduct(ctx context.Context, productID, profileID int) error {
	_, err := s.entClient.ProductShippingProfile.Create().
		SetProductID(productID).
		SetProfileID(profileID).
		Save(ctx)
	return err
}
