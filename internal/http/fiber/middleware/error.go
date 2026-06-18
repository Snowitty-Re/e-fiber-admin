package middleware

import (
	"github.com/gofiber/fiber/v3"

	pkgerr "github.com/Snowitty-Re/e-fiber-admin/internal/pkg/errors"
)

func ErrorHandler(c fiber.Ctx, err error) error {
	if ae, ok := pkgerr.As(err); ok {
		return c.Status(ae.Status).JSON(errorBody(ae))
	}
	return c.Status(fiber.StatusInternalServerError).JSON(map[string]any{
		"error": map[string]any{
			"code":       "INTERNAL_ERROR",
			"message":    "internal server error",
			"status":     fiber.StatusInternalServerError,
			"request_id": c.Locals("requestid", ""),
		},
	})
}

func errorBody(ae *pkgerr.AppError) map[string]any {
	body := map[string]any{
		"code":    ae.Code,
		"message": ae.Message,
		"status":  ae.Status,
	}
	if len(ae.Details) > 0 {
		body["details"] = ae.Details
	}
	return map[string]any{"error": body}
}
