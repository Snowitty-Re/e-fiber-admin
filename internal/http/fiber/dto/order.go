package dto

type CheckoutRequest struct {
	CartID           int            `json:"cart_id" validate:"required"`
	Email            string         `json:"email" validate:"required"`
	ShippingAddress  map[string]any `json:"shipping_address"`
	BillingAddress   map[string]any `json:"billing_address,omitempty"`
	ShippingOptionID int            `json:"shipping_option_id,omitempty"`
	DiscountCode     string         `json:"discount_code,omitempty"`
}

type OrderItemResponse struct {
	ID          int    `json:"id"`
	VariantID   int    `json:"variant_id,omitempty"`
	SKU         string `json:"sku"`
	Title       string `json:"title"`
	Quantity    int    `json:"quantity"`
	UnitAmount  int64  `json:"unit_amount"`
	TotalAmount int64  `json:"total_amount"`
}

type FulfillmentItemRequest struct {
	OrderItemID int `json:"order_item_id"`
	Quantity    int `json:"quantity"`
}

type CreateFulfillmentRequest struct {
	TrackingNumber string                   `json:"tracking_number,omitempty"`
	Items          []FulfillmentItemRequest `json:"items"`
}

type ReturnItemRequest struct {
	OrderItemID int    `json:"order_item_id"`
	Quantity    int    `json:"quantity"`
	Reason      string `json:"reason,omitempty"`
}

type CreateReturnRequest struct {
	Reason       string              `json:"reason"`
	RefundAmount int64               `json:"refund_amount"`
	CurrencyCode string              `json:"currency_code,omitempty"`
	Items        []ReturnItemRequest `json:"items"`
}

type OrderResponse struct {
	ID                int                 `json:"id"`
	Number            string              `json:"number"`
	Email             string              `json:"email"`
	CustomerID        int                 `json:"customer_id,omitempty"`
	CurrencyCode      string              `json:"currency_code"`
	Status            string              `json:"status"`
	FulfillmentStatus string              `json:"fulfillment_status"`
	PaymentStatus     string              `json:"payment_status"`
	Totals            map[string]any      `json:"totals,omitempty"`
	ShippingAddress   map[string]any      `json:"shipping_address,omitempty"`
	PlacedAt          *string             `json:"placed_at,omitempty"`
	Items             []OrderItemResponse `json:"items"`
}
