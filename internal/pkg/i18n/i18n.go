package i18n

import (
	"strings"

	"github.com/gofiber/fiber/v3"
)

type Locale struct {
	Code string
}

func FromHeader(accept string, defaultLocale string) string {
	if accept == "" {
		return defaultLocale
	}
	parts := strings.Split(accept, ",")
	for _, p := range parts {
		tag := strings.TrimSpace(strings.Split(p, ";")[0])
		if tag == "" || tag == "*" {
			continue
		}
		return normalize(tag)
	}
	return defaultLocale
}

func FromRequest(c fiber.Ctx, defaultLocale string) string {
	return FromHeader(c.Get("Accept-Language"), defaultLocale)
}

func normalize(tag string) string {
	tag = strings.ToLower(strings.TrimSpace(tag))
	if i := strings.Index(tag, "-"); i >= 0 {
		base := tag[:i]
		region := strings.ToLower(tag[i+1:])
		if len(region) == 2 {
			return base + "-" + strings.ToUpper(region)
		}
		return base
	}
	return tag
}

func IsValid(code string) bool {
	code = strings.TrimSpace(code)
	if code == "" {
		return false
	}
	parts := strings.Split(code, "-")
	if len(parts) < 1 || len(parts[0]) != 2 {
		return false
	}
	return true
}
