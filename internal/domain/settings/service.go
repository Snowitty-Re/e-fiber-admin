package settings

import (
	"context"
	"fmt"

	"github.com/Snowitty-Re/e-fiber-admin/internal/ent"
	"github.com/Snowitty-Re/e-fiber-admin/internal/ent/store"
	pkgerr "github.com/Snowitty-Re/e-fiber-admin/internal/pkg/errors"
)

type Service struct {
	entClient *ent.Client
}

func NewService(entClient *ent.Client) *Service {
	return &Service{entClient: entClient}
}

func (s *Service) GetStore(ctx context.Context) (*ent.Store, error) {
	st, err := s.entClient.Store.Query().First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, pkgerr.ErrNotFound
		}
		return nil, fmt.Errorf("query store: %w", err)
	}
	return st, nil
}

type StoreUpdate struct {
	Name            *string
	Slug            *string
	SiteType        *string
	DefaultLocale   *string
	DefaultCurrency *string
	Timezone        *string
	Status          *string
	FeatureFlags    map[string]bool
}

func (s *Service) UpdateStore(ctx context.Context, id int, in StoreUpdate) (*ent.Store, error) {
	st, err := s.entClient.Store.Query().Where(store.IDEQ(id)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, pkgerr.ErrNotFound
		}
		return nil, fmt.Errorf("query store: %w", err)
	}
	b := st.Update()
	if in.Name != nil {
		b.SetName(*in.Name)
	}
	if in.Slug != nil {
		b.SetSlug(*in.Slug)
	}
	if in.SiteType != nil {
		b.SetSiteType(store.SiteType(*in.SiteType))
	}
	if in.DefaultLocale != nil {
		b.SetDefaultLocale(*in.DefaultLocale)
	}
	if in.DefaultCurrency != nil {
		b.SetDefaultCurrency(*in.DefaultCurrency)
	}
	if in.Timezone != nil {
		b.SetTimezone(*in.Timezone)
	}
	if in.Status != nil {
		b.SetStatus(store.Status(*in.Status))
	}
	if in.FeatureFlags != nil {
		b.SetFeatureFlags(in.FeatureFlags)
	}
	return b.Save(ctx)
}

func (s *Service) UpdateFeatureFlags(ctx context.Context, id int, flags map[string]bool) (*ent.Store, error) {
	st, err := s.entClient.Store.Query().Where(store.IDEQ(id)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, pkgerr.ErrNotFound
		}
		return nil, fmt.Errorf("query store: %w", err)
	}
	current := st.FeatureFlags
	if current == nil {
		current = map[string]bool{}
	}
	for k, v := range flags {
		current[k] = v
	}
	return st.Update().SetFeatureFlags(current).Save(ctx)
}
