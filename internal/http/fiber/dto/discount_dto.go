package dto

type DiscountRuleResponse struct {
	ID         int    `json:"id"`
	Type       string `json:"type"`
	Value      int64  `json:"value"`
	Allocation string `json:"allocation"`
}

type DiscountResponse struct {
	ID         int                    `json:"id"`
	Code       string                 `json:"code,omitempty"`
	Name       string                 `json:"name"`
	Status     string                 `json:"status"`
	StartsAt   *string                `json:"starts_at,omitempty"`
	EndsAt     *string                `json:"ends_at,omitempty"`
	UsageLimit *int                   `json:"usage_limit,omitempty"`
	UsageCount int                    `json:"usage_count"`
	Rules      []DiscountRuleResponse `json:"rules"`
}

type CreateDiscountRequest struct {
	Code       string            `json:"code,omitempty"`
	Name       string            `json:"name" validate:"required"`
	StartsAt   *string           `json:"starts_at,omitempty"`
	EndsAt     *string           `json:"ends_at,omitempty"`
	UsageLimit int               `json:"usage_limit,omitempty"`
	Rules      []DiscountRuleReq `json:"rules" validate:"required,min=1"`
}

type DiscountRuleReq struct {
	Type       string `json:"type" validate:"required"`
	Value      int64  `json:"value"`
	Allocation string `json:"allocation,omitempty"`
}

type ValidateCodeRequest struct {
	Code string `json:"code" validate:"required"`
}
