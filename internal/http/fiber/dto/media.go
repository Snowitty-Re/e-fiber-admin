package dto

type MediaResponse struct {
	ID        int    `json:"id"`
	Key       string `json:"key"`
	URL       string `json:"url"`
	MimeType  string `json:"mime_type"`
	SizeBytes int64  `json:"size_bytes"`
	Width     int    `json:"width,omitempty"`
	Height    int    `json:"height,omitempty"`
	Kind      string `json:"kind"`
}

type MediaListResponse struct {
	Data       []MediaResponse `json:"data"`
	Pagination Pagination      `json:"pagination"`
}

type Pagination struct {
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Total    int64 `json:"total"`
	HasMore  bool  `json:"has_more"`
}
