package dto

type StoreResponse struct {
	ID               int            `json:"id"`
	Name             string         `json:"name"`
	Slug             string         `json:"slug"`
	SiteType         string         `json:"site_type"`
	DefaultLocale    string         `json:"default_locale"`
	DefaultCurrency  string         `json:"default_currency"`
	FeatureFlags     map[string]bool `json:"feature_flags"`
	Timezone         string         `json:"timezone"`
	Status           string         `json:"status"`
}

type UpdateStoreRequest struct {
	Name            *string `json:"name,omitempty"`
	Slug            *string `json:"slug,omitempty"`
	SiteType        *string `json:"site_type,omitempty"`
	DefaultLocale   *string `json:"default_locale,omitempty"`
	DefaultCurrency *string `json:"default_currency,omitempty"`
	Timezone        *string `json:"timezone,omitempty"`
	Status          *string `json:"status,omitempty"`
}

type UpdateFeatureFlagsRequest struct {
	Flags map[string]bool `json:"flags"`
}
