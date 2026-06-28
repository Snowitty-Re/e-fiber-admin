package dto

type ShippingProfileResponse struct {
	ID          int                      `json:"id"`
	Name        string                   `json:"name"`
	ProductType string                   `json:"product_type"`
	Options     []ShippingOptionResponse `json:"options"`
}

type ShippingOptionResponse struct {
	ID            int    `json:"id"`
	ProfileID     int    `json:"profile_id"`
	Name          string `json:"name"`
	PriceAmount   int64  `json:"price_amount"`
	PriceCurrency string `json:"price_currency"`
	EstimatedDays int    `json:"estimated_days"`
	IsActive      bool   `json:"is_active"`
}

type CreateShippingProfileRequest struct {
	Name        string `json:"name" validate:"required"`
	ProductType string `json:"product_type,omitempty"`
}

type CreateShippingOptionRequest struct {
	ProfileID     int    `json:"profile_id" validate:"required"`
	Name          string `json:"name" validate:"required"`
	PriceAmount   int64  `json:"price_amount"`
	PriceCurrency string `json:"price_currency,omitempty"`
	EstimatedDays int64  `json:"estimated_days,omitempty"`
	IsActive      bool   `json:"is_active"`
}
