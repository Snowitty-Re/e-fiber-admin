package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"

	discsvc "github.com/Snowitty-Re/e-fiber-admin/internal/domain/discount"
	"github.com/Snowitty-Re/e-fiber-admin/internal/ent"
	"github.com/Snowitty-Re/e-fiber-admin/internal/http/fiber/dto"
	pkgerr "github.com/Snowitty-Re/e-fiber-admin/internal/pkg/errors"
)

type DiscountHandler struct {
	discountService *discsvc.Service
}

func NewDiscountHandler(ds *discsvc.Service) *DiscountHandler {
	return &DiscountHandler{discountService: ds}
}

func (h *DiscountHandler) AdminCreate(c fiber.Ctx) error {
	var req dto.CreateDiscountRequest
	if err := c.Bind().Body(&req); err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	if req.Name == "" || len(req.Rules) == 0 {
		return pkgerr.ErrValidation.WithDetails(
			pkgerr.FieldError{Field: "name", Issue: "required"},
			pkgerr.FieldError{Field: "rules", Issue: "at least one required"},
		)
	}

	in := discsvc.DiscountInput{Code: req.Code, Name: req.Name}
	if req.StartsAt != nil {
		if t, err := time.Parse(time.RFC3339, *req.StartsAt); err == nil {
			in.StartsAt = &t
		}
	}
	if req.EndsAt != nil {
		if t, err := time.Parse(time.RFC3339, *req.EndsAt); err == nil {
			in.EndsAt = &t
		}
	}
	in.UsageLimit = req.UsageLimit
	for _, r := range req.Rules {
		alloc := r.Allocation
		if alloc == "" {
			alloc = "all"
		}
		in.Rules = append(in.Rules, discsvc.RuleInput{
			Type: r.Type, Value: r.Value, Allocation: alloc,
		})
	}

	d, err := h.discountService.Create(c.Context(), in)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"discount": toDiscountResponse(d)})
}

func (h *DiscountHandler) AdminList(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))
	items, total, err := h.discountService.List(c.Context(), page, pageSize)
	if err != nil {
		return err
	}
	var data []dto.DiscountResponse
	for _, d := range items {
		data = append(data, toDiscountResponse(d))
	}
	return c.JSON(fiber.Map{
		"data": data,
		"pagination": fiber.Map{"page": page, "page_size": pageSize, "total": total,
			"has_more": int64(page*pageSize) < int64(total)},
	})
}

func (h *DiscountHandler) ValidateCode(c fiber.Ctx) error {
	var req dto.ValidateCodeRequest
	if err := c.Bind().Body(&req); err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	if req.Code == "" {
		return pkgerr.ErrValidation.WithDetails(pkgerr.FieldError{Field: "code", Issue: "required"})
	}
	d, err := h.discountService.ValidateCode(c.Context(), req.Code)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"discount": toDiscountResponse(d)})
}

func toDiscountResponse(d *ent.Discount) dto.DiscountResponse {
	resp := dto.DiscountResponse{
		ID: d.ID, Code: d.Code, Name: d.Name, Status: string(d.Status),
		UsageCount: d.UsageCount,
		Rules:      []dto.DiscountRuleResponse{},
	}
	if d.UsageLimit > 0 {
		ul := d.UsageLimit
		resp.UsageLimit = &ul
	}
	if d.StartsAt != nil {
		resp.StartsAt = &[]string{d.StartsAt.Format(time.RFC3339)}[0]
	}
	if d.EndsAt != nil {
		resp.EndsAt = &[]string{d.EndsAt.Format(time.RFC3339)}[0]
	}
	for _, r := range d.Edges.Rules {
		resp.Rules = append(resp.Rules, dto.DiscountRuleResponse{
			ID: r.ID, Type: string(r.Type), Value: r.Value, Allocation: string(r.Allocation),
		})
	}
	return resp
}
