package dto

type CartItemResponse struct {
	ID          int    `json:"id"`
	VariantID   int    `json:"variant_id"`
	ProductID   int    `json:"product_id"`
	SKU         string `json:"sku"`
	Quantity    int    `json:"quantity"`
	UnitAmount  int64  `json:"unit_amount"`
	TotalAmount int64  `json:"total_amount"`
	Currency    string `json:"currency_code"`
}

type CartResponse struct {
	ID          int                `json:"id"`
	Status      string             `json:"status"`
	Currency    string             `json:"currency_code"`
	Locale      string             `json:"locale"`
	Items       []CartItemResponse `json:"items"`
	TotalAmount int64              `json:"total_amount"`
}

type AddCartItemRequest struct {
	VariantID int `json:"variant_id" validate:"required"`
	Quantity  int `json:"quantity" validate:"required,min=1"`
}

type UpdateCartItemRequest struct {
	Quantity int `json:"quantity" validate:"required,min=1"`
}

type CreateCartRequest struct {
	CurrencyCode string `json:"currency_code,omitempty"`
	Locale       string `json:"locale,omitempty"`
}
