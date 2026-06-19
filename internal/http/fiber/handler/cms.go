package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"

	"github.com/Snowitty-Re/e-fiber-admin/internal/domain/cms"
	"github.com/Snowitty-Re/e-fiber-admin/internal/ent"
	entblogpost "github.com/Snowitty-Re/e-fiber-admin/internal/ent/blogpost"
	entpage "github.com/Snowitty-Re/e-fiber-admin/internal/ent/page"
	"github.com/Snowitty-Re/e-fiber-admin/internal/http/fiber/dto"
	pkgerr "github.com/Snowitty-Re/e-fiber-admin/internal/pkg/errors"
	"github.com/Snowitty-Re/e-fiber-admin/internal/pkg/i18n"
)

type CMSHandler struct {
	cmsService *cms.Service
	entClient  *ent.Client
}

func NewCMSHandler(cmsService *cms.Service, entClient *ent.Client) *CMSHandler {
	return &CMSHandler{cmsService: cmsService, entClient: entClient}
}

func (h *CMSHandler) CreatePage(c fiber.Ctx) error {
	var req dto.CreatePageRequest
	if err := c.Bind().Body(&req); err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	if req.Slug == "" || len(req.Translations) == 0 {
		return pkgerr.ErrValidation.WithDetails(
			pkgerr.FieldError{Field: "slug", Issue: "required"},
			pkgerr.FieldError{Field: "translations", Issue: "at least one required"},
		)
	}
	tpl := req.Template
	if tpl == "" {
		tpl = "default"
	}
	in := cms.PageInput{Slug: req.Slug, Template: tpl}
	for _, t := range req.Translations {
		in.Translations = append(in.Translations, cms.PageTranslationInput{
			Locale: t.Locale, Title: t.Title, Content: t.Content,
			SeoTitle: t.SeoTitle, SeoDesc: t.SeoDesc,
		})
	}
	p, err := h.cmsService.CreatePage(c.Context(), in)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"page": toPageResponse(p)})
}

func (h *CMSHandler) ListPages(c fiber.Ctx) error {
	page_, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))
	items, total, err := h.cmsService.ListPages(c.Context(), page_, pageSize)
	if err != nil {
		return err
	}
	var data []dto.PageResponse
	for _, p := range items {
		data = append(data, toPageResponse(p))
	}
	return c.JSON(fiber.Map{
		"data": data,
		"pagination": fiber.Map{"page": page_, "page_size": pageSize, "total": total,
			"has_more": int64(page_*pageSize) < int64(total)},
	})
}

func (h *CMSHandler) GetPage(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	p, err := h.cmsService.GetPage(c.Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"page": toPageResponse(p)})
}

func (h *CMSHandler) PublishPage(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	if err := h.cmsService.PublishPage(c.Context(), id); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *CMSHandler) DeletePage(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	if err := h.cmsService.DeletePage(c.Context(), id); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *CMSHandler) CreateBlogPost(c fiber.Ctx) error {
	var req dto.CreateBlogPostRequest
	if err := c.Bind().Body(&req); err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	if req.Slug == "" || len(req.Translations) == 0 {
		return pkgerr.ErrValidation.WithDetails(
			pkgerr.FieldError{Field: "slug", Issue: "required"},
			pkgerr.FieldError{Field: "translations", Issue: "at least one required"},
		)
	}
	adminID, _ := c.Locals("admin_id").(int64)
	in := cms.BlogPostInput{Slug: req.Slug, AuthorAdminID: int(adminID)}
	for _, t := range req.Translations {
		in.Translations = append(in.Translations, cms.BlogPostTranslationInput{
			Locale: t.Locale, Title: t.Title, Excerpt: t.Excerpt,
			Content: t.Content, SeoTitle: t.SeoTitle, SeoDesc: t.SeoDesc,
		})
	}
	bp, err := h.cmsService.CreateBlogPost(c.Context(), in)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"blog_post": toBlogPostResponse(bp)})
}

func (h *CMSHandler) ListBlogPosts(c fiber.Ctx) error {
	page_, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))
	items, total, err := h.cmsService.ListBlogPosts(c.Context(), page_, pageSize)
	if err != nil {
		return err
	}
	var data []dto.BlogPostResponse
	for _, bp := range items {
		data = append(data, toBlogPostResponse(bp))
	}
	return c.JSON(fiber.Map{
		"data": data,
		"pagination": fiber.Map{"page": page_, "page_size": pageSize, "total": total,
			"has_more": int64(page_*pageSize) < int64(total)},
	})
}

func (h *CMSHandler) PublishBlogPost(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	if err := h.cmsService.PublishBlogPost(c.Context(), id); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *CMSHandler) DeleteBlogPost(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return pkgerr.ErrBadRequest.WithCause(err)
	}
	if err := h.cmsService.DeleteBlogPost(c.Context(), id); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *CMSHandler) StoreGetPage(c fiber.Ctx) error {
	store, err := h.entClient.Store.Query().First(c.Context())
	if err != nil {
		return pkgerr.ErrServiceUnavailable.WithCause(err)
	}
	locale := i18n.FromRequest(c, store.DefaultLocale)
	slug := c.Params("slug")
	p, err := h.cmsService.GetPageBySlug(c.Context(), slug)
	if err != nil {
		return err
	}
	if p.Status != entpage.StatusPublished {
		return pkgerr.ErrNotFound
	}
	resp := dto.PageResponse{
		ID: p.ID, Slug: p.Slug, Status: string(p.Status), Template: p.Template,
	}
	for _, t := range p.Edges.Translations {
		if t.Locale == locale {
			resp.Translations = []dto.PageTranslationResponse{{
				Locale: t.Locale, Title: t.Title, Content: t.Content,
				SeoTitle: t.SeoTitle, SeoDesc: t.SeoDesc,
			}}
			break
		}
	}
	if len(resp.Translations) == 0 && len(p.Edges.Translations) > 0 {
		t := p.Edges.Translations[0]
		resp.Translations = []dto.PageTranslationResponse{{
			Locale: t.Locale, Title: t.Title, Content: t.Content,
			SeoTitle: t.SeoTitle, SeoDesc: t.SeoDesc,
		}}
	}
	return c.JSON(fiber.Map{"page": resp})
}

func (h *CMSHandler) StoreListBlogPosts(c fiber.Ctx) error {
	page_, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))
	store, err := h.entClient.Store.Query().First(c.Context())
	if err != nil {
		return pkgerr.ErrServiceUnavailable.WithCause(err)
	}
	locale := i18n.FromRequest(c, store.DefaultLocale)

	items, total, err := h.cmsService.ListBlogPosts(c.Context(), page_, pageSize)
	if err != nil {
		return err
	}
	var data []dto.BlogPostResponse
	for _, bp := range items {
		if bp.Status != entblogpost.StatusPublished {
			continue
		}
		resp := dto.BlogPostResponse{ID: bp.ID, Slug: bp.Slug, Status: string(bp.Status)}
		for _, t := range bp.Edges.Translations {
			if t.Locale == locale {
				resp.Translations = []dto.BlogPostTranslationResponse{{
					Locale: t.Locale, Title: t.Title, Excerpt: t.Excerpt,
					Content: t.Content, SeoTitle: t.SeoTitle, SeoDesc: t.SeoDesc,
				}}
				break
			}
		}
		data = append(data, resp)
	}
	return c.JSON(fiber.Map{
		"data": data,
		"pagination": fiber.Map{"page": page_, "page_size": pageSize, "total": total,
			"has_more": int64(page_*pageSize) < int64(total)},
	})
}

func toPageResponse(p *ent.Page) dto.PageResponse {
	resp := dto.PageResponse{
		ID: p.ID, Slug: p.Slug, Status: string(p.Status), Template: p.Template,
	}
	for _, t := range p.Edges.Translations {
		resp.Translations = append(resp.Translations, dto.PageTranslationResponse{
			Locale: t.Locale, Title: t.Title, Content: t.Content,
			SeoTitle: t.SeoTitle, SeoDesc: t.SeoDesc,
		})
	}
	return resp
}

func toBlogPostResponse(bp *ent.BlogPost) dto.BlogPostResponse {
	resp := dto.BlogPostResponse{
		ID: bp.ID, Slug: bp.Slug, Status: string(bp.Status),
	}
	for _, t := range bp.Edges.Translations {
		resp.Translations = append(resp.Translations, dto.BlogPostTranslationResponse{
			Locale: t.Locale, Title: t.Title, Excerpt: t.Excerpt,
			Content: t.Content, SeoTitle: t.SeoTitle, SeoDesc: t.SeoDesc,
		})
	}
	return resp
}
