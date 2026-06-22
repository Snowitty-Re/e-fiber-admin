package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"

	"github.com/Snowitty-Re/e-fiber-admin/internal/domain/inquiry"
	"github.com/Snowitty-Re/e-fiber-admin/internal/ent"
	"github.com/Snowitty-Re/e-fiber-admin/internal/http/fiber/dto"
	pkgerr "github.com/Snowitty-Re/e-fiber-admin/internal/pkg/errors"
)

type InquiryHandler struct {
	inquiryService *inquiry.Service
}

func NewInquiryHandler(is *inquiry.Service) *InquiryHandler {
	return &InquiryHandler{inquiryService: is}
}

func (h *InquiryHandler) CreateForm(c fiber.Ctx) error {
	var req dto.CreateFormRequest
	if err := c.Bind().Body(&req); err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	if req.Slug == "" || len(req.Translations) == 0 {
		return pkgerr.ErrValidation.WithDetails(
			pkgerr.FieldError{Field: "slug", Issue: "required"},
			pkgerr.FieldError{Field: "translations", Issue: "at least one required"},
		)
	}
	in := inquiry.FormInput{
		Slug: req.Slug, Fields: req.Fields, NotifyEmails: req.NotifyEmails,
		IsActive: req.IsActive,
	}
	for _, t := range req.Translations {
		in.Translations = append(in.Translations, inquiry.FormTranslationInput{
			Locale: t.Locale, Title: t.Title, FieldLabels: t.FieldLabels,
		})
	}
	f, err := h.inquiryService.CreateForm(c.Context(), in)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"form": toFormResponse(f)})
}

func (h *InquiryHandler) ListForms(c fiber.Ctx) error {
	forms, err := h.inquiryService.ListForms(c.Context())
	if err != nil {
		return err
	}
	var data []dto.FormResponse
	for _, f := range forms {
		data = append(data, toFormResponse(f))
	}
	return c.JSON(fiber.Map{"data": data})
}

func (h *InquiryHandler) AdminListInquiries(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))
	status := c.Query("status", "")
	items, total, err := h.inquiryService.ListInquiries(c.Context(), page, pageSize, status)
	if err != nil {
		return err
	}
	var data []dto.InquiryResponse
	for _, inq := range items {
		data = append(data, toInquiryResponse(inq))
	}
	return c.JSON(fiber.Map{
		"data": data,
		"pagination": fiber.Map{"page": page, "page_size": pageSize, "total": total,
			"has_more": int64(page*pageSize) < int64(total)},
	})
}

func (h *InquiryHandler) AdminGetInquiry(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	inq, err := h.inquiryService.GetInquiry(c.Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"inquiry": toInquiryResponse(inq)})
}

func (h *InquiryHandler) AdminAssign(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	var req dto.AssignInquiryRequest
	if err := c.Bind().Body(&req); err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	if err := h.inquiryService.AssignInquiry(c.Context(), id, req.AdminID); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *InquiryHandler) AdminUpdateStatus(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	var req dto.UpdateInquiryStatusRequest
	if err := c.Bind().Body(&req); err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	if err := h.inquiryService.UpdateStatus(c.Context(), id, req.Status); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *InquiryHandler) StoreSubmitInquiry(c fiber.Ctx) error {
	var req dto.SubmitInquiryRequest
	if err := c.Bind().Body(&req); err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	if req.FormSlug == "" || req.Email == "" {
		return pkgerr.ErrValidation.WithDetails(
			pkgerr.FieldError{Field: "form_slug", Issue: "required"},
			pkgerr.FieldError{Field: "email", Issue: "required"},
		)
	}
	customerID, _ := c.Locals("customer_id").(int64)
	in := inquiry.InquirySubmitInput{
		FormSlug: req.FormSlug, Email: req.Email, Phone: req.Phone,
		Name: req.Name, Company: req.Company, Payload: req.Payload,
		ProductID: req.ProductID,
	}
	if customerID > 0 {
		in.CustomerID = int(customerID)
	}
	inq, err := h.inquiryService.SubmitInquiry(c.Context(), in)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"inquiry": toInquiryResponse(inq)})
}

func toFormResponse(f *ent.FormDefinition) dto.FormResponse {
	resp := dto.FormResponse{
		ID: f.ID, Slug: f.Slug, IsActive: f.IsActive,
		Fields: f.Fields, NotifyEmails: f.NotifyEmails,
	}
	for _, t := range f.Edges.Translations {
		resp.Translations = append(resp.Translations, dto.FormTranslationResp{
			Locale: t.Locale, Title: t.Title, FieldLabels: t.FieldLabels,
		})
	}
	return resp
}

func toInquiryResponse(inq *ent.Inquiry) dto.InquiryResponse {
	return dto.InquiryResponse{
		ID: inq.ID, FormID: inq.FormID, Email: inq.Email, Phone: inq.Phone,
		Name: inq.Name, Company: inq.Company, Payload: inq.Payload,
		Status: string(inq.Status), ProductID: inq.ProductID,
		AssignedAdminID: inq.AssignedAdminID,
	}
}