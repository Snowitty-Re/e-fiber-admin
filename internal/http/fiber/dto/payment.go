package dto

type PaymentProviderResponse struct {
	ID       int    `json:"id"`
	Code     string `json:"code"`
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
}

type CreatePaymentProviderRequest struct {
	Code   string         `json:"code" validate:"required"`
	Name   string         `json:"name" validate:"required"`
	Config map[string]any `json:"config,omitempty"`
}

type CreatePaymentSessionRequest struct {
	OrderID      int    `json:"order_id" validate:"required"`
	ProviderCode string `json:"provider_code" validate:"required"`
}

type PaymentSessionResponse struct {
	ID           int    `json:"id"`
	OrderID      int    `json:"order_id"`
	ProviderCode string `json:"provider_code"`
	Status       string `json:"status"`
	Amount       int64  `json:"amount"`
	CurrencyCode string `json:"currency_code"`
}

type AuthorizeRequest struct {
	SessionID int `json:"session_id" validate:"required"`
}

type CaptureRequest struct {
	SessionID int   `json:"session_id" validate:"required"`
	Amount    int64 `json:"amount" validate:"required"`
}
