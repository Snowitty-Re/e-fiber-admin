package handler

import (
	"github.com/gofiber/fiber/v3"

	authsvc "github.com/Snowitty-Re/e-fiber-admin/internal/domain/auth"
	"github.com/Snowitty-Re/e-fiber-admin/internal/http/fiber/dto"
	pkgerr "github.com/Snowitty-Re/e-fiber-admin/internal/pkg/errors"
)

type AuthHandler struct {
	authService *authsvc.Service
}

func NewAuthHandler(authService *authsvc.Service) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Login(c fiber.Ctx) error {
	var req dto.LoginRequest
	if err := c.Bind().Body(&req); err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	if req.Email == "" || req.Password == "" {
		return pkgerr.ErrValidation.WithDetails(
			pkgerr.FieldError{Field: "email", Issue: "required"},
			pkgerr.FieldError{Field: "password", Issue: "required"},
		)
	}
	identity, pair, err := h.authService.Login(c.Context(), req.Email, req.Password)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"access_token":  pair.AccessToken,
		"refresh_token": pair.RefreshToken,
		"expires_in":    pair.ExpiresIn,
		"admin": dto.AdminProfile{
			ID:        identity.ID,
			Email:     identity.Email,
			FirstName: identity.FirstName,
			LastName:  identity.LastName,
			Roles:     identity.Roles,
			Perms:     identity.Perms,
		},
	})
}

func (h *AuthHandler) Refresh(c fiber.Ctx) error {
	var req dto.RefreshRequest
	if err := c.Bind().Body(&req); err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	if req.RefreshToken == "" {
		return pkgerr.ErrValidation.WithDetails(
			pkgerr.FieldError{Field: "refresh_token", Issue: "required"},
		)
	}
	pair, err := h.authService.Refresh(c.Context(), req.RefreshToken)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{
		"access_token":  pair.AccessToken,
		"refresh_token": pair.RefreshToken,
		"expires_in":    pair.ExpiresIn,
	})
}

func (h *AuthHandler) Logout(c fiber.Ctx) error {
	var req dto.LogoutRequest
	_ = c.Bind().Body(&req)
	_ = h.authService.Logout(c.Context(), req.RefreshToken)
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *AuthHandler) Me(c fiber.Ctx) error {
	adminID, ok := c.Locals("admin_id").(int64)
	if !ok {
		return pkgerr.ErrUnauthorized
	}
	identity, err := h.authService.Me(c.Context(), adminID)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{
		"admin": dto.AdminProfile{
			ID:        identity.ID,
			Email:     identity.Email,
			FirstName: identity.FirstName,
			LastName:  identity.LastName,
			Roles:     identity.Roles,
			Perms:     identity.Perms,
		},
	})
}
