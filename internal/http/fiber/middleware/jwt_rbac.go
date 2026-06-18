package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v3"

	authsvc "github.com/Snowitty/e-fiber-admin/internal/domain/auth"
	pkgerr "github.com/Snowitty/e-fiber-admin/internal/pkg/errors"
)

func extractToken(c fiber.Ctx) string {
	auth := c.Get("Authorization")
	if auth == "" {
		return ""
	}
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return strings.TrimSpace(parts[1])
	}
	return ""
}

func JWTAuth(authService *authsvc.Service) fiber.Handler {
	return func(c fiber.Ctx) error {
		token := extractToken(c)
		if token == "" {
			return pkgerr.ErrUnauthorized
		}
		claims, err := authService.ParseAccess(token)
		if err != nil {
			return err
		}
		c.Locals("admin_id", claims.AdminID)
		c.Locals("admin_roles", claims.Roles)
		c.Locals("admin_perms", claims.Perms)
		return c.Next()
	}
}

func RBAC(required ...string) fiber.Handler {
	return func(c fiber.Ctx) error {
		perms, ok := c.Locals("admin_perms").([]string)
		if !ok {
			return pkgerr.ErrForbidden
		}
		permSet := make(map[string]bool, len(perms))
		for _, p := range perms {
			permSet[p] = true
		}
		for _, r := range required {
			if !permSet[r] {
				return pkgerr.ErrForbidden
			}
		}
		return c.Next()
	}
}

func RBACAny(required ...string) fiber.Handler {
	return func(c fiber.Ctx) error {
		perms, ok := c.Locals("admin_perms").([]string)
		if !ok {
			return pkgerr.ErrForbidden
		}
		permSet := make(map[string]bool, len(perms))
		for _, p := range perms {
			permSet[p] = true
		}
		for _, r := range required {
			if permSet[r] {
				return c.Next()
			}
		}
		return pkgerr.ErrForbidden
	}
}
