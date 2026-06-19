package dto

type PageTranslationRequest struct {
	Locale   string         `json:"locale"`
	Title    string         `json:"title"`
	Content  map[string]any `json:"content,omitempty"`
	SeoTitle string         `json:"seo_title,omitempty"`
	SeoDesc  string         `json:"seo_desc,omitempty"`
}

type CreatePageRequest struct {
	Slug         string                   `json:"slug"`
	Template     string                   `json:"template,omitempty"`
	Translations []PageTranslationRequest `json:"translations"`
}

type PageResponse struct {
	ID           int                       `json:"id"`
	Slug         string                    `json:"slug"`
	Status       string                    `json:"status"`
	Template     string                    `json:"template"`
	Translations []PageTranslationResponse `json:"translations"`
}

type PageTranslationResponse struct {
	Locale   string         `json:"locale"`
	Title    string         `json:"title"`
	Content  map[string]any `json:"content,omitempty"`
	SeoTitle string         `json:"seo_title,omitempty"`
	SeoDesc  string         `json:"seo_desc,omitempty"`
}

type BlogPostTranslationRequest struct {
	Locale   string `json:"locale"`
	Title    string `json:"title"`
	Excerpt  string `json:"excerpt,omitempty"`
	Content  string `json:"content,omitempty"`
	SeoTitle string `json:"seo_title,omitempty"`
	SeoDesc  string `json:"seo_desc,omitempty"`
}

type CreateBlogPostRequest struct {
	Slug         string                       `json:"slug"`
	Translations []BlogPostTranslationRequest `json:"translations"`
}

type BlogPostResponse struct {
	ID           int                           `json:"id"`
	Slug         string                        `json:"slug"`
	Status       string                        `json:"status"`
	Translations []BlogPostTranslationResponse `json:"translations"`
}

type BlogPostTranslationResponse struct {
	Locale   string `json:"locale"`
	Title    string `json:"title"`
	Excerpt  string `json:"excerpt,omitempty"`
	Content  string `json:"content,omitempty"`
	SeoTitle string `json:"seo_title,omitempty"`
	SeoDesc  string `json:"seo_desc,omitempty"`
}
