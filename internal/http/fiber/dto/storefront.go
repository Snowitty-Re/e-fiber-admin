package dto

type StoreProductResponse struct {
	ID          int                  `json:"id"`
	Slug        string               `json:"slug"`
	ProductType string               `json:"product_type"`
	Locale      string               `json:"locale"`
	Currency    string               `json:"currency"`
	Title       string               `json:"title"`
	Subtitle    string               `json:"subtitle,omitempty"`
	Description string               `json:"description,omitempty"`
	SeoTitle    string               `json:"seo_title,omitempty"`
	SeoDesc     string               `json:"seo_desc,omitempty"`
	Variants    []StoreVariantResponse `json:"variants"`
}

type StoreVariantResponse struct {
	ID             int   `json:"id"`
	SKU            string `json:"sku"`
	Inventory      int   `json:"inventory"`
	AllowBackorder bool  `json:"allow_backorder"`
	Price          int64 `json:"price"`
	CompareAtPrice int64 `json:"compare_at_price,omitempty"`
}
