package cms

import (
	"context"
	"fmt"
	"time"

	"github.com/Snowitty-Re/e-fiber-admin/internal/ent"
	"github.com/Snowitty-Re/e-fiber-admin/internal/ent/blogpost"
	"github.com/Snowitty-Re/e-fiber-admin/internal/ent/menu"
	"github.com/Snowitty-Re/e-fiber-admin/internal/ent/menuitem"
	"github.com/Snowitty-Re/e-fiber-admin/internal/ent/page"
	pkgerr "github.com/Snowitty-Re/e-fiber-admin/internal/pkg/errors"
)

type Service struct {
	entClient *ent.Client
}

func NewService(entClient *ent.Client) *Service {
	return &Service{entClient: entClient}
}

type PageTranslationInput struct {
	Locale   string
	Title    string
	Content  map[string]any
	SeoTitle string
	SeoDesc  string
}

type PageInput struct {
	Slug         string
	Template     string
	Translations []PageTranslationInput
}

func (s *Service) CreatePage(ctx context.Context, in PageInput) (*ent.Page, error) {
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	p, err := tx.Page.Create().
		SetSlug(in.Slug).
		SetTemplate(in.Template).
		SetStatus(page.StatusDraft).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create page: %w", err)
	}
	for _, t := range in.Translations {
		_, err = tx.PageTranslation.Create().
			SetPageID(p.ID).
			SetLocale(t.Locale).
			SetTitle(t.Title).
			SetContent(t.Content).
			SetSeoTitle(t.SeoTitle).
			SetSeoDesc(t.SeoDesc).
			Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("create page translation: %w", err)
		}
	}
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return s.GetPage(ctx, p.ID)
}

func (s *Service) GetPage(ctx context.Context, id int) (*ent.Page, error) {
	p, err := s.entClient.Page.Query().
		Where(page.IDEQ(id)).
		WithTranslations().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, pkgerr.ErrNotFound
		}
		return nil, fmt.Errorf("query page: %w", err)
	}
	return p, nil
}

func (s *Service) GetPageBySlug(ctx context.Context, slug string) (*ent.Page, error) {
	p, err := s.entClient.Page.Query().
		Where(page.SlugEQ(slug)).
		WithTranslations().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, pkgerr.ErrNotFound
		}
		return nil, fmt.Errorf("query page: %w", err)
	}
	return p, nil
}

func (s *Service) ListPages(ctx context.Context, page_ int, pageSize int) ([]*ent.Page, int, error) {
	q := s.entClient.Page.Query().WithTranslations()
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	if page_ < 1 {
		page_ = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	items, err := q.Order(ent.Desc(page.FieldID)).Limit(pageSize).Offset((page_ - 1) * pageSize).All(ctx)
	return items, total, err
}

func (s *Service) PublishPage(ctx context.Context, id int) error {
	return s.entClient.Page.UpdateOneID(id).
		SetStatus(page.StatusPublished).
		SetPublishedAt(time.Now()).
		Exec(ctx)
}

func (s *Service) DeletePage(ctx context.Context, id int) error {
	return s.entClient.Page.DeleteOneID(id).Exec(ctx)
}

type BlogPostTranslationInput struct {
	Locale   string
	Title    string
	Excerpt  string
	Content  string
	SeoTitle string
	SeoDesc  string
}

type BlogPostInput struct {
	Slug          string
	AuthorAdminID int
	Translations  []BlogPostTranslationInput
}

func (s *Service) CreateBlogPost(ctx context.Context, in BlogPostInput) (*ent.BlogPost, error) {
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	bp, err := tx.BlogPost.Create().
		SetSlug(in.Slug).
		SetAuthorAdminID(in.AuthorAdminID).
		SetStatus(blogpost.StatusDraft).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create blog post: %w", err)
	}
	for _, t := range in.Translations {
		_, err = tx.BlogPostTranslation.Create().
			SetBlogPostID(bp.ID).
			SetLocale(t.Locale).
			SetTitle(t.Title).
			SetExcerpt(t.Excerpt).
			SetContent(t.Content).
			SetSeoTitle(t.SeoTitle).
			SetSeoDesc(t.SeoDesc).
			Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("create blog translation: %w", err)
		}
	}
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return s.GetBlogPost(ctx, bp.ID)
}

func (s *Service) GetBlogPost(ctx context.Context, id int) (*ent.BlogPost, error) {
	bp, err := s.entClient.BlogPost.Query().
		Where(blogpost.IDEQ(id)).
		WithTranslations().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, pkgerr.ErrNotFound
		}
		return nil, fmt.Errorf("query blog post: %w", err)
	}
	return bp, nil
}

func (s *Service) GetBlogPostBySlug(ctx context.Context, slug string) (*ent.BlogPost, error) {
	bp, err := s.entClient.BlogPost.Query().
		Where(blogpost.SlugEQ(slug)).
		WithTranslations().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, pkgerr.ErrNotFound
		}
		return nil, fmt.Errorf("query blog post: %w", err)
	}
	return bp, nil
}

func (s *Service) ListBlogPosts(ctx context.Context, page_ int, pageSize int) ([]*ent.BlogPost, int, error) {
	q := s.entClient.BlogPost.Query().WithTranslations()
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	if page_ < 1 {
		page_ = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	items, err := q.Order(ent.Desc(blogpost.FieldID)).Limit(pageSize).Offset((page_ - 1) * pageSize).All(ctx)
	return items, total, err
}

func (s *Service) PublishBlogPost(ctx context.Context, id int) error {
	return s.entClient.BlogPost.UpdateOneID(id).
		SetStatus(blogpost.StatusPublished).
		SetPublishedAt(time.Now()).
		Exec(ctx)
}

func (s *Service) DeleteBlogPost(ctx context.Context, id int) error {
	return s.entClient.BlogPost.DeleteOneID(id).Exec(ctx)
}

func (s *Service) GetMenuByLocation(ctx context.Context, location, locale string) (*ent.Menu, error) {
	m, err := s.entClient.Menu.Query().
		Where(menu.LocationEQ(location)).
		WithItems(func(q *ent.MenuItemQuery) {
			q.WithTranslations().Order(ent.Asc(menuitem.FieldPosition))
		}).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, pkgerr.ErrNotFound
		}
		return nil, fmt.Errorf("query menu: %w", err)
	}
	return m, nil
}
