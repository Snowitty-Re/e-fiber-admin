package region

import (
	"context"
	"fmt"

	"github.com/Snowitty-Re/e-fiber-admin/internal/ent"
	"github.com/Snowitty-Re/e-fiber-admin/internal/ent/currency"
	"github.com/Snowitty-Re/e-fiber-admin/internal/ent/locale"
	"github.com/Snowitty-Re/e-fiber-admin/internal/ent/region"
	"github.com/Snowitty-Re/e-fiber-admin/internal/ent/taxrate"
	pkgerr "github.com/Snowitty-Re/e-fiber-admin/internal/pkg/errors"
)

type Service struct {
	entClient *ent.Client
}

func NewService(entClient *ent.Client) *Service {
	return &Service{entClient: entClient}
}

func (s *Service) ListLocales(ctx context.Context) ([]*ent.Locale, error) {
	return s.entClient.Locale.Query().Where(locale.IsActiveEQ(true)).All(ctx)
}

func (s *Service) CreateLocale(ctx context.Context, code, name string) (*ent.Locale, error) {
	return s.entClient.Locale.Create().SetCode(code).SetName(name).SetIsActive(true).Save(ctx)
}

func (s *Service) ListCurrencies(ctx context.Context) ([]*ent.Currency, error) {
	return s.entClient.Currency.Query().Where(currency.IsActiveEQ(true)).All(ctx)
}

func (s *Service) CreateCurrency(ctx context.Context, code, name, symbol string, precision int) (*ent.Currency, error) {
	return s.entClient.Currency.Create().
		SetCode(code).SetName(name).SetSymbol(symbol).SetPrecision(precision).SetIsActive(true).
		Save(ctx)
}

type RegionInput struct {
	Name         string
	Locale       string
	CurrencyCode string
	TaxInclusive bool
	Countries    []string
}

func (s *Service) ListRegions(ctx context.Context) ([]*ent.Region, error) {
	return s.entClient.Region.Query().WithTaxRates().All(ctx)
}

func (s *Service) GetRegion(ctx context.Context, id int) (*ent.Region, error) {
	r, err := s.entClient.Region.Query().Where(region.IDEQ(id)).WithTaxRates().Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, pkgerr.ErrNotFound
		}
		return nil, fmt.Errorf("query region: %w", err)
	}
	return r, nil
}

func (s *Service) CreateRegion(ctx context.Context, in RegionInput) (*ent.Region, error) {
	return s.entClient.Region.Create().
		SetName(in.Name).
		SetLocale(in.Locale).
		SetCurrencyCode(in.CurrencyCode).
		SetTaxInclusive(in.TaxInclusive).
		SetCountries(in.Countries).
		Save(ctx)
}

func (s *Service) UpdateRegion(ctx context.Context, id int, in RegionInput) (*ent.Region, error) {
	r, err := s.entClient.Region.Query().Where(region.IDEQ(id)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, pkgerr.ErrNotFound
		}
		return nil, fmt.Errorf("query region: %w", err)
	}
	return r.Update().
		SetName(in.Name).
		SetLocale(in.Locale).
		SetCurrencyCode(in.CurrencyCode).
		SetTaxInclusive(in.TaxInclusive).
		SetCountries(in.Countries).
		Save(ctx)
}

func (s *Service) DeleteRegion(ctx context.Context, id int) error {
	return s.entClient.Region.DeleteOneID(id).Exec(ctx)
}

type TaxRateInput struct {
	RegionID    int
	CountryCode string
	Rate        float64
	Name        string
	Priority    int
}

func (s *Service) ListTaxRates(ctx context.Context, regionID int) ([]*ent.TaxRate, error) {
	q := s.entClient.TaxRate.Query()
	if regionID > 0 {
		q = q.Where(taxrate.RegionIDEQ(regionID))
	}
	return q.All(ctx)
}

func (s *Service) CreateTaxRate(ctx context.Context, in TaxRateInput) (*ent.TaxRate, error) {
	return s.entClient.TaxRate.Create().
		SetRegionID(in.RegionID).
		SetCountryCode(in.CountryCode).
		SetRate(in.Rate).
		SetName(in.Name).
		SetPriority(in.Priority).
		Save(ctx)
}

func (s *Service) DeleteTaxRate(ctx context.Context, id int) error {
	return s.entClient.TaxRate.DeleteOneID(id).Exec(ctx)
}
