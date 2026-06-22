package dto

type CustomerRegisterRequest struct {
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required,min=6"`
	FirstName       string `json:"first_name,omitempty"`
	LastName        string `json:"last_name,omitempty"`
	DefaultCurrency string `json:"default_currency,omitempty"`
	DefaultLocale   string `json:"default_locale,omitempty"`
}

type CustomerLoginRequest struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type CustomerRefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type CustomerProfile struct {
	ID              int    `json:"id"`
	Email           string `json:"email"`
	FirstName       string `json:"first_name"`
	LastName        string `json:"last_name"`
	DefaultCurrency string `json:"default_currency"`
	DefaultLocale   string `json:"default_locale"`
}

type CustomerTokenResponse struct {
	AccessToken  string          `json:"access_token"`
	RefreshToken string          `json:"refresh_token"`
	ExpiresIn    int64           `json:"expires_in"`
	Customer     CustomerProfile `json:"customer"`
}

type CustomerListResponse struct {
	Data       []CustomerAdminResponse `json:"data"`
	Pagination Pagination              `json:"pagination"`
}

type CustomerAdminResponse struct {
	ID     int    `json:"id"`
	Email  string `json:"email"`
	Status string `json:"status"`
}
