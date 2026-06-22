package dto

type FormFieldDef struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required,omitempty"`
	Label    string `json:"label,omitempty"`
}

type FormTranslationRequest struct {
	Locale      string            `json:"locale"`
	Title       string            `json:"title"`
	FieldLabels map[string]string `json:"field_labels,omitempty"`
}

type CreateFormRequest struct {
	Slug         string                  `json:"slug"`
	Fields       []map[string]any        `json:"fields,omitempty"`
	NotifyEmails []string               `json:"notify_emails,omitempty"`
	IsActive     bool                   `json:"is_active"`
	Translations []FormTranslationRequest `json:"translations"`
}

type FormResponse struct {
	ID           int                    `json:"id"`
	Slug         string                 `json:"slug"`
	IsActive     bool                   `json:"is_active"`
	Fields       []map[string]any       `json:"fields,omitempty"`
	NotifyEmails []string               `json:"notify_emails,omitempty"`
	Translations []FormTranslationResp  `json:"translations"`
}

type FormTranslationResp struct {
	Locale      string            `json:"locale"`
	Title       string            `json:"title"`
	FieldLabels map[string]string `json:"field_labels,omitempty"`
}

type SubmitInquiryRequest struct {
	FormSlug  string         `json:"form_slug"`
	Email     string         `json:"email"`
	Phone     string         `json:"phone,omitempty"`
	Name      string         `json:"name,omitempty"`
	Company   string         `json:"company,omitempty"`
	Payload   map[string]any `json:"payload"`
	ProductID int            `json:"product_id,omitempty"`
}

type InquiryResponse struct {
	ID        int            `json:"id"`
	FormID    int            `json:"form_id"`
	Email     string         `json:"email"`
	Phone     string         `json:"phone,omitempty"`
	Name      string         `json:"name,omitempty"`
	Company   string         `json:"company,omitempty"`
	Payload   map[string]any `json:"payload,omitempty"`
	Status    string         `json:"status"`
	ProductID int            `json:"product_id,omitempty"`
	AssignedAdminID int      `json:"assigned_admin_id,omitempty"`
}

type AssignInquiryRequest struct {
	AdminID int `json:"admin_id"`
}

type UpdateInquiryStatusRequest struct {
	Status string `json:"status"`
}