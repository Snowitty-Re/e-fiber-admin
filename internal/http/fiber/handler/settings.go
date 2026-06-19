package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"

	"github.com/Snowitty-Re/e-fiber-admin/internal/domain/settings"
	"github.com/Snowitty-Re/e-fiber-admin/internal/ent"
	"github.com/Snowitty-Re/e-fiber-admin/internal/http/fiber/dto"
	pkgerr "github.com/Snowitty-Re/e-fiber-admin/internal/pkg/errors"
)

type SettingsHandler struct {
	settingsService *settings.Service
}

func NewSettingsHandler(settingsService *settings.Service) *SettingsHandler {
	return &SettingsHandler{settingsService: settingsService}
}

func (h *SettingsHandler) GetStore(c fiber.Ctx) error {
	st, err := h.settingsService.GetStore(c.Context())
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"store": toStoreResponse(st)})
}

func (h *SettingsHandler) UpdateStore(c fiber.Ctx) error {
	var req dto.UpdateStoreRequest
	if err := c.Bind().Body(&req); err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	st, err := h.settingsService.GetStore(c.Context())
	if err != nil {
		return err
	}
	in := settings.StoreUpdate{
		Name: req.Name, Slug: req.Slug, SiteType: req.SiteType,
		DefaultLocale: req.DefaultLocale, DefaultCurrency: req.DefaultCurrency,
		Timezone: req.Timezone, Status: req.Status,
	}
	updated, err := h.settingsService.UpdateStore(c.Context(), st.ID, in)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"store": toStoreResponse(updated)})
}

func (h *SettingsHandler) UpdateFeatureFlags(c fiber.Ctx) error {
	var req dto.UpdateFeatureFlagsRequest
	if err := c.Bind().Body(&req); err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	st, err := h.settingsService.UpdateFeatureFlags(c.Context(), id, req.Flags)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"store": toStoreResponse(st)})
}

func toStoreResponse(st *ent.Store) dto.StoreResponse {
	return dto.StoreResponse{
		ID: st.ID, Name: st.Name, Slug: st.Slug,
		SiteType:      string(st.SiteType),
		DefaultLocale: st.DefaultLocale, DefaultCurrency: st.DefaultCurrency,
		FeatureFlags: st.FeatureFlags, Timezone: st.Timezone,
		Status: string(st.Status),
	}
}
