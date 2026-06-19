package dto

type LocaleResponse struct {
	ID       int    `json:"id"`
	Code     string `json:"code"`
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
}

type CurrencyResponse struct {
	ID        int    `json:"id"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	Symbol    string `json:"symbol"`
	Precision int    `json:"precision"`
	IsActive  bool   `json:"is_active"`
}

type CreateLocaleRequest struct {
	Code string `json:"code" validate:"required,max=8"`
	Name string `json:"name" validate:"required"`
}

type CreateCurrencyRequest struct {
	Code      string `json:"code" validate:"required,len=3"`
	Name      string `json:"name" validate:"required"`
	Symbol    string `json:"symbol"`
	Precision int    `json:"precision"`
}

type RegionResponse struct {
	ID           int      `json:"id"`
	Name         string   `json:"name"`
	Locale       string   `json:"locale"`
	CurrencyCode string   `json:"currency_code"`
	TaxInclusive bool     `json:"tax_inclusive"`
	Countries    []string `json:"countries"`
	Status       string   `json:"status"`
}

type CreateRegionRequest struct {
	Name         string   `json:"name" validate:"required"`
	Locale       string   `json:"locale" validate:"required"`
	CurrencyCode string   `json:"currency_code" validate:"required,len=3"`
	TaxInclusive bool     `json:"tax_inclusive"`
	Countries    []string `json:"countries"`
}

type TaxRateResponse struct {
	ID          int     `json:"id"`
	RegionID    int     `json:"region_id"`
	CountryCode string  `json:"country_code"`
	Rate        float64 `json:"rate"`
	Name        string  `json:"name"`
	Priority    int     `json:"priority"`
}

type CreateTaxRateRequest struct {
	RegionID    int     `json:"region_id" validate:"required"`
	CountryCode string  `json:"country_code"`
	Rate        float64 `json:"rate" validate:"required"`
	Name        string  `json:"name" validate:"required"`
	Priority    int     `json:"priority"`
}
