package pagination

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
)

const (
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 100
)

type Params struct {
	Page     int
	PageSize int
	Sort     string
	Cursor   string
	Expand   []string
	Fields   []string
}

type Response struct {
	Data       []any `json:"data"`
	Pagination Meta  `json:"pagination"`
}

type Meta struct {
	Page       int    `json:"page"`
	PageSize   int    `json:"page_size"`
	Total      int64  `json:"total"`
	HasMore    bool   `json:"has_more"`
	NextCursor string `json:"next_cursor,omitempty"`
}

func Parse(c fiber.Ctx) Params {
	p := Params{
		Page:     intQuery(c, "page", DefaultPage),
		PageSize: intQuery(c, "page_size", DefaultPageSize),
		Sort:     c.Query("sort", ""),
		Cursor:   c.Query("cursor", ""),
	}
	if p.Page < 1 {
		p.Page = DefaultPage
	}
	if p.PageSize < 1 {
		p.PageSize = DefaultPageSize
	}
	if p.PageSize > MaxPageSize {
		p.PageSize = MaxPageSize
	}
	if expand := c.Query("expand", ""); expand != "" {
		p.Expand = strings.Split(expand, ",")
	}
	if fields := c.Query("fields", ""); fields != "" {
		p.Fields = strings.Split(fields, ",")
	}
	return p
}

func (p Params) Offset() int { return (p.Page - 1) * p.PageSize }

func MetaFrom(p Params, total int64) Meta {
	return Meta{
		Page:     p.Page,
		PageSize: p.PageSize,
		Total:    total,
		HasMore:  int64(p.Page*p.PageSize) < total,
	}
}

func intQuery(c fiber.Ctx, key string, fallback int) int {
	v := c.Query(key, "")
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
