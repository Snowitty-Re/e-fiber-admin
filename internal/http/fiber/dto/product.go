package dto

type TranslationRequest struct {
	Locale      string `json:"locale"`
	Title       string `json:"title"`
	Subtitle    string `json:"subtitle,omitempty"`
	Description string `json:"description,omitempty"`
	Material    string `json:"material,omitempty"`
	Origin      string `json:"origin,omitempty"`
	Packing     string `json:"packing,omitempty"`
	SeoTitle    string `json:"seo_title,omitempty"`
	SeoDesc     string `json:"seo_desc,omitempty"`
}

type VariantPriceRequest struct {
	CurrencyCode    string `json:"currency_code"`
	Amount          int64  `json:"amount"`
	CompareAtAmount int64  `json:"compare_at_amount,omitempty"`
}

type VariantRequest struct {
	SKU            string                `json:"sku"`
	Barcode        string                `json:"barcode,omitempty"`
	WeightG        int                   `json:"weight_g,omitempty"`
	AllowBackorder bool                  `json:"allow_backorder"`
	Inventory      int                   `json:"inventory"`
	Position       int                   `json:"position,omitempty"`
	Prices         []VariantPriceRequest `json:"prices,omitempty"`
}

type CreateProductRequest struct {
	Slug           string               `json:"slug" validate:"required"`
	ProductType    string               `json:"product_type"`
	CategoryID     int                  `json:"category_id,omitempty"`
	WeightG        int                  `json:"weight_g,omitempty"`
	IsVirtual      bool                 `json:"is_virtual"`
	IsDownloadable bool                 `json:"is_downloadable"`
	Translations   []TranslationRequest `json:"translations" validate:"required,min=1"`
	Variants       []VariantRequest     `json:"variants" validate:"required,min=1"`
}

type TranslationResponse struct {
	Locale      string `json:"locale"`
	Title       string `json:"title"`
	Subtitle    string `json:"subtitle,omitempty"`
	Description string `json:"description,omitempty"`
	Material    string `json:"material,omitempty"`
	Origin      string `json:"origin,omitempty"`
	Packing     string `json:"packing,omitempty"`
	SeoTitle    string `json:"seo_title,omitempty"`
	SeoDesc     string `json:"seo_desc,omitempty"`
}

type VariantPriceResponse struct {
	CurrencyCode    string `json:"currency_code"`
	Amount          int64  `json:"amount"`
	CompareAtAmount int64  `json:"compare_at_amount,omitempty"`
}

type VariantResponse struct {
	ID             int                    `json:"id"`
	SKU            string                 `json:"sku"`
	Barcode        string                 `json:"barcode,omitempty"`
	WeightG        int                    `json:"weight_g,omitempty"`
	AllowBackorder bool                   `json:"allow_backorder"`
	Inventory      int                    `json:"inventory"`
	Position       int                    `json:"position"`
	Prices         []VariantPriceResponse `json:"prices,omitempty"`
}

type ProductResponse struct {
	ID             int                    `json:"id"`
	Slug           string                 `json:"slug"`
	ProductType    string                 `json:"product_type"`
	Status         string                 `json:"status"`
	CategoryID     int                    `json:"category_id,omitempty"`
	WeightG        int                    `json:"weight_g,omitempty"`
	IsVirtual      bool                   `json:"is_virtual"`
	IsDownloadable bool                   `json:"is_downloadable"`
	Translations   []TranslationResponse  `json:"translations"`
	Variants       []VariantResponse      `json:"variants"`
}
