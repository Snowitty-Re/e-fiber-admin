package discount

import (
	"context"
	"fmt"
	"time"

	"github.com/Snowitty-Re/e-fiber-admin/internal/ent"
	"github.com/Snowitty-Re/e-fiber-admin/internal/ent/discount"
	"github.com/Snowitty-Re/e-fiber-admin/internal/ent/discountrule"
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

type RuleInput struct {
	Type       string
	Value      int64
	Allocation string
	Target     map[string]any
}

type DiscountInput struct {
	Code       string
	Name       string
	StartsAt   *time.Time
	EndsAt     *time.Time
	UsageLimit int
	Rules      []RuleInput
}

func (s *Service) Create(ctx context.Context, in DiscountInput) (*ent.Discount, error) {
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	b := tx.Discount.Create().
		SetName(in.Name).
		SetStatus(discount.StatusActive)
	if in.Code != "" {
		b.SetCode(in.Code)
	}
	if in.StartsAt != nil {
		b.SetStartsAt(*in.StartsAt)
	}
	if in.EndsAt != nil {
		b.SetEndsAt(*in.EndsAt)
	}
	if in.UsageLimit > 0 {
		b.SetUsageLimit(in.UsageLimit)
	}

	d, err := b.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create discount: %w", err)
	}

	for _, r := range in.Rules {
		_, err = tx.DiscountRule.Create().
			SetDiscountID(d.ID).
			SetType(discountrule.Type(r.Type)).
			SetValue(r.Value).
			SetAllocation(discountrule.Allocation(r.Allocation)).
			SetTarget(r.Target).
			Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("create rule: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return s.Get(ctx, d.ID)
}

func (s *Service) Get(ctx context.Context, id int) (*ent.Discount, error) {
	d, err := s.entClient.Discount.Query().
		Where(discount.IDEQ(id)).
		WithRules().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, pkgerr.ErrNotFound
		}
		return nil, fmt.Errorf("query discount: %w", err)
	}
	return d, nil
}

func (s *Service) List(ctx context.Context, page, pageSize int) ([]*ent.Discount, int, error) {
	q := s.entClient.Discount.Query().WithRules()
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
	items, err := q.Order(ent.Desc(discount.FieldID)).Limit(pageSize).Offset((page - 1) * pageSize).All(ctx)
	return items, total, err
}

func (s *Service) ValidateCode(ctx context.Context, code string) (*ent.Discount, error) {
	d, err := s.entClient.Discount.Query().
		Where(discount.CodeEQ(code), discount.StatusEQ(discount.StatusActive)).
		WithRules().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, pkgerr.New("DISCOUNT_NOT_FOUND", 404, "discount code not found or inactive")
		}
		return nil, fmt.Errorf("query discount: %w", err)
	}

	now := time.Now()
	if d.StartsAt != nil && now.Before(*d.StartsAt) {
		return nil, pkgerr.New("DISCOUNT_NOT_STARTED", 409, "discount not yet started")
	}
	if d.EndsAt != nil && now.After(*d.EndsAt) {
		return nil, pkgerr.New("DISCOUNT_EXPIRED", 409, "discount expired")
	}
	if d.UsageLimit > 0 && d.UsageCount >= d.UsageLimit {
		return nil, pkgerr.New("DISCOUNT.Usage_LIMIT_REACHED", 409, "usage limit reached")
	}

	return d, nil
}

func (s *Service) IncrementUsage(ctx context.Context, id int) error {
	d, err := s.Get(ctx, id)
	if err != nil {
		return err
	}
	_, err = d.Update().SetUsageCount(d.UsageCount + 1).Save(ctx)
	if err != nil {
		return err
	}
	if s.bus != nil {
		_ = s.bus.PublishSimple(ctx, "discount.redeemed", "discount", fmt.Sprintf("%d", id), map[string]any{
			"discount_id": id, "usage_count": d.UsageCount + 1,
		})
	}
	return nil
}

func (s *Service) ApplyToTotal(rules []*ent.DiscountRule, subtotal int64) (int64, string) {
	if len(rules) == 0 {
		return 0, ""
	}
	r := rules[0]
	switch r.Type {
	case discountrule.TypePercentage:
		discountAmount := subtotal * r.Value / 100
		return discountAmount, "percentage"
	case discountrule.TypeFixed:
		if r.Value > subtotal {
			return subtotal, "fixed"
		}
		return r.Value, "fixed"
	case discountrule.TypeShipping:
		return 0, "shipping"
	default:
		return 0, ""
	}
}

func (s *Service) ExpireOverdue(ctx context.Context) (int, error) {
	now := time.Now()
	items, err := s.entClient.Discount.Query().
		Where(discount.StatusEQ(discount.StatusActive)).
		WithRules().
		All(ctx)
	if err != nil {
		return 0, err
	}
	expired := 0
	for _, d := range items {
		if d.EndsAt != nil && now.After(*d.EndsAt) {
			_, _ = d.Update().SetStatus(discount.StatusExpired).Save(ctx)
			expired++
			if s.bus != nil {
				_ = s.bus.PublishSimple(ctx, "discount.expired", "discount", fmt.Sprintf("%d", d.ID), nil)
			}
		}
	}
	return expired, nil
}
