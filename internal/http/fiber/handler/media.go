package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"

	mediabc "github.com/Snowitty-Re/e-fiber-admin/internal/domain/media"
	"github.com/Snowitty-Re/e-fiber-admin/internal/ent"
	"github.com/Snowitty-Re/e-fiber-admin/internal/http/fiber/dto"
	pkgerr "github.com/Snowitty-Re/e-fiber-admin/internal/pkg/errors"
)

type MediaHandler struct {
	mediaService *mediabc.Service
}

func NewMediaHandler(mediaService *mediabc.Service) *MediaHandler {
	return &MediaHandler{mediaService: mediaService}
}

func (h *MediaHandler) Upload(c fiber.Ctx) error {
	fh, err := c.FormFile("file")
	if err != nil {
		return pkgerr.ErrValidation.WithDetails(
			pkgerr.FieldError{Field: "file", Issue: "required"},
		)
	}
	result, err := h.mediaService.Upload(c.Context(), fh)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"media": result})
}

func (h *MediaHandler) List(c fiber.Ctx) error {
	kind := c.Query("kind", "")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	limit := pageSize
	offset := (page - 1) * pageSize

	items, total, err := h.mediaService.List(c.Context(), kind, limit, offset)
	if err != nil {
		return err
	}

	var data []dto.MediaResponse
	for _, m := range items {
		data = append(data, toMediaResponse(m))
	}
	return c.JSON(dto.MediaListResponse{
		Data: data,
		Pagination: dto.Pagination{
			Page: page, PageSize: pageSize, Total: int64(total),
			HasMore: int64(offset+limit) < int64(total),
		},
	})
}

func (h *MediaHandler) Get(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	m, err := h.mediaService.Get(c.Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"media": toMediaResponse(m)})
}

func (h *MediaHandler) Delete(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	if err := h.mediaService.Delete(c.Context(), id); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func toMediaResponse(m *ent.Media) dto.MediaResponse {
	return dto.MediaResponse{
		ID: m.ID, Key: m.Key, URL: m.URL, MimeType: m.MimeType,
		SizeBytes: m.SizeBytes, Width: m.Width, Height: m.Height,
		Kind: string(m.Kind),
	}
}
