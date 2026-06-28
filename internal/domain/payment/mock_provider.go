package payment

import (
	"context"
	"fmt"

	"github.com/Snowitty-Re/e-fiber-admin/internal/ent"
)

type MockProvider struct{}

func (MockProvider) Code() string { return "mock" }

func (MockProvider) Authorize(ctx context.Context, session *ent.PaymentSession) (map[string]any, error) {
	return map[string]any{
		"mock": true, "session_id": session.ID, "token": fmt.Sprintf("mock_auth_%d", session.ID),
	}, nil
}

func (MockProvider) Capture(ctx context.Context, session *ent.PaymentSession, amount int64) error {
	return nil
}

func (MockProvider) Refund(ctx context.Context, session *ent.PaymentSession, amount int64) error {
	return nil
}
