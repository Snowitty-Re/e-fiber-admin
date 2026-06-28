package payment

import (
	"context"
	"fmt"

	"github.com/Snowitty-Re/e-fiber-admin/internal/ent"
	"github.com/Snowitty-Re/e-fiber-admin/internal/ent/paymentprovider"
	"github.com/Snowitty-Re/e-fiber-admin/internal/ent/paymentsession"
	"github.com/Snowitty-Re/e-fiber-admin/internal/events"
	pkgerr "github.com/Snowitty-Re/e-fiber-admin/internal/pkg/errors"
)

type Provider interface {
	Code() string
	Authorize(ctx context.Context, session *ent.PaymentSession) (map[string]any, error)
	Capture(ctx context.Context, session *ent.PaymentSession, amount int64) error
	Refund(ctx context.Context, session *ent.PaymentSession, amount int64) error
}

type Service struct {
	entClient *ent.Client
	providers map[string]Provider
	bus       *events.Bus
}

func NewService(entClient *ent.Client, bus *events.Bus) *Service {
	return &Service{
		entClient: entClient,
		providers: make(map[string]Provider),
		bus:       bus,
	}
}

func (s *Service) RegisterProvider(p Provider) {
	s.providers[p.Code()] = p
}

func (s *Service) ListProviders(ctx context.Context) ([]*ent.PaymentProvider, error) {
	return s.entClient.PaymentProvider.Query().All(ctx)
}

func (s *Service) CreateProvider(ctx context.Context, code, name string, config map[string]any) (*ent.PaymentProvider, error) {
	return s.entClient.PaymentProvider.Create().
		SetCode(code).
		SetName(name).
		SetIsActive(true).
		SetConfig(config).
		Save(ctx)
}

func (s *Service) CreateSession(ctx context.Context, orderID int, providerCode, currencyCode string, amount int64) (*ent.PaymentSession, error) {
	_, err := s.entClient.PaymentProvider.Query().
		Where(paymentprovider.CodeEQ(providerCode), paymentprovider.IsActiveEQ(true)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, pkgerr.New("PAYMENT_PROVIDER_NOT_FOUND", 404, "provider not found or inactive")
		}
		return nil, fmt.Errorf("query provider: %w", err)
	}

	return s.entClient.PaymentSession.Create().
		SetOrderID(orderID).
		SetProviderCode(providerCode).
		SetStatus(paymentsession.StatusPending).
		SetAmount(amount).
		SetCurrencyCode(currencyCode).
		Save(ctx)
}

func (s *Service) Authorize(ctx context.Context, sessionID int) (*ent.PaymentSession, error) {
	session, err := s.entClient.PaymentSession.Query().
		Where(paymentsession.IDEQ(sessionID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, pkgerr.ErrNotFound
		}
		return nil, fmt.Errorf("query session: %w", err)
	}

	provider, ok := s.providers[session.ProviderCode]
	if !ok {
		return nil, pkgerr.New("PAYMENT_PROVIDER_NOT_REGISTERED", 503, "provider "+session.ProviderCode+" not registered")
	}

	providerData, err := provider.Authorize(ctx, session)
	if err != nil {
		_, _ = session.Update().SetStatus(paymentsession.StatusFailed).Save(ctx)
		return nil, fmt.Errorf("authorize: %w", err)
	}

	updated, err := session.Update().
		SetStatus(paymentsession.StatusAuthorized).
		SetProviderData(providerData).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("update session: %w", err)
	}

	_, _ = s.entClient.Transaction.Create().
		SetPaymentSessionID(sessionID).
		SetAmount(session.Amount).
		SetCurrencyCode(session.CurrencyCode).
		SetType("authorize").
		SetStatus("succeeded").
		Save(ctx)

	if s.bus != nil {
		_ = s.bus.PublishSimple(ctx, "payment.authorized", "payment", fmt.Sprintf("%d", sessionID), map[string]any{
			"session_id": sessionID, "order_id": session.OrderID,
		})
	}
	return updated, nil
}

func (s *Service) Capture(ctx context.Context, sessionID int, amount int64) error {
	session, err := s.entClient.PaymentSession.Query().
		Where(paymentsession.IDEQ(sessionID)).
		Only(ctx)
	if err != nil {
		return pkgerr.ErrNotFound
	}

	provider, ok := s.providers[session.ProviderCode]
	if !ok {
		return pkgerr.New("PAYMENT_PROVIDER_NOT_REGISTERED", 503, "provider not registered")
	}

	if err := provider.Capture(ctx, session, amount); err != nil {
		_, _ = s.entClient.Transaction.Create().
			SetPaymentSessionID(sessionID).
			SetAmount(amount).
			SetCurrencyCode(session.CurrencyCode).
			SetType("capture").
			SetStatus("failed").
			Save(ctx)
		return fmt.Errorf("capture: %w", err)
	}

	_, _ = session.Update().SetStatus(paymentsession.StatusCaptured).Save(ctx)
	_, _ = s.entClient.Transaction.Create().
		SetPaymentSessionID(sessionID).
		SetAmount(amount).
		SetCurrencyCode(session.CurrencyCode).
		SetType("capture").
		SetStatus("succeeded").
		Save(ctx)

	if s.bus != nil {
		_ = s.bus.PublishSimple(ctx, "payment.captured", "payment", fmt.Sprintf("%d", sessionID), map[string]any{
			"session_id": sessionID, "order_id": session.OrderID, "amount": amount,
		})
	}
	return nil
}

func (s *Service) Refund(ctx context.Context, sessionID int, amount int64) error {
	session, err := s.entClient.PaymentSession.Query().
		Where(paymentsession.IDEQ(sessionID)).
		Only(ctx)
	if err != nil {
		return pkgerr.ErrNotFound
	}

	provider, ok := s.providers[session.ProviderCode]
	if !ok {
		return pkgerr.New("PAYMENT_PROVIDER_NOT_REGISTERED", 503, "provider not registered")
	}

	if err := provider.Refund(ctx, session, amount); err != nil {
		return fmt.Errorf("refund: %w", err)
	}

	_, _ = s.entClient.Transaction.Create().
		SetPaymentSessionID(sessionID).
		SetAmount(amount).
		SetCurrencyCode(session.CurrencyCode).
		SetType("refund").
		SetStatus("succeeded").
		Save(ctx)

	if s.bus != nil {
		_ = s.bus.PublishSimple(ctx, "payment.refunded", "payment", fmt.Sprintf("%d", sessionID), map[string]any{
			"session_id": sessionID, "order_id": session.OrderID, "amount": amount,
		})
	}
	return nil
}

func (s *Service) GetSession(ctx context.Context, id int) (*ent.PaymentSession, error) {
	session, err := s.entClient.PaymentSession.Query().
		Where(paymentsession.IDEQ(id)).
		WithTransactions().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, pkgerr.ErrNotFound
		}
		return nil, fmt.Errorf("query session: %w", err)
	}
	return session, nil
}

func (s *Service) GetSessionByOrder(ctx context.Context, orderID int) (*ent.PaymentSession, error) {
	session, err := s.entClient.PaymentSession.Query().
		Where(paymentsession.OrderIDEQ(orderID)).
		WithTransactions().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, pkgerr.ErrNotFound
		}
		return nil, fmt.Errorf("query session by order: %w", err)
	}
	return session, nil
}
