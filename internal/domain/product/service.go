package product

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Snowitty-Re/e-fiber-admin/internal/ent"
	"github.com/Snowitty-Re/e-fiber-admin/internal/ent/product"
	"github.com/Snowitty-Re/e-fiber-admin/internal/ent/producttranslation"
	"github.com/Snowitty-Re/e-fiber-admin/internal/ent/variant"
	"github.com/Snowitty-Re/e-fiber-admin/internal/ent/variantprice"
	pkgerr "github.com/Snowitty-Re/e-fiber-admin/internal/pkg/errors"
)

type Service struct {
	entClient *ent.Client
	db        *sql.DB
}

func NewService(entClient *ent.Client, db *sql.DB) *Service {
	return &Service{entClient: entClient, db: db}
}

type TranslationInput struct {
	Locale     string
	Title      string
	Subtitle   string
	Description string
	Material   string
	Origin     string
	Packing    string
	SeoTitle   string
	SeoDesc    string
}

type VariantPriceInput struct {
	CurrencyCode     string
	Amount           int64
	CompareAtAmount  int64
}

type VariantInput struct {
	SKU           string
	Barcode       string
	WeightG       int
	AllowBackorder bool
	Inventory     int
	Position      int
	Prices        []VariantPriceInput
}

type ProductInput struct {
	Slug          string
	ProductType   string
	CategoryID    int
	WeightG       int
	IsVirtual     bool
	IsDownloadable bool
	Translations  []TranslationInput
	Variants      []VariantInput
}

func (s *Service) Create(ctx context.Context, in ProductInput) (*ent.Product, error) {
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	pType := product.ProductType(in.ProductType)
	if pType == "" {
		pType = product.ProductTypeSimple
	}

	pb := tx.Product.Create().
		SetSlug(in.Slug).
		SetProductType(pType).
		SetStatus(product.StatusDraft).
		SetIsVirtual(in.IsVirtual).
		SetIsDownloadable(in.IsDownloadable)
	if in.CategoryID > 0 {
		pb.SetCategoryID(in.CategoryID)
	}
	if in.WeightG > 0 {
		pb.SetWeightG(in.WeightG)
	}

	p, err := pb.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create product: %w", err)
	}

	for _, t := range in.Translations {
		_, err = tx.ProductTranslation.Create().
			SetProductID(p.ID).
			SetLocale(t.Locale).
			SetTitle(t.Title).
			SetSubtitle(t.Subtitle).
			SetDescription(t.Description).
			SetMaterial(t.Material).
			SetOrigin(t.Origin).
			SetPacking(t.Packing).
			SetSeoTitle(t.SeoTitle).
			SetSeoDesc(t.SeoDesc).
			Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("create translation: %w", err)
		}
	}

	for _, v := range in.Variants {
		vb := tx.Variant.Create().
			SetProductID(p.ID).
			SetSku(v.SKU).
			SetBarcode(v.Barcode).
			SetAllowBackorder(v.AllowBackorder).
			SetInventory(v.Inventory).
			SetPosition(v.Position)
		if v.WeightG > 0 {
			vb.SetWeightG(v.WeightG)
		}
		vEnt, err := vb.Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("create variant: %w", err)
		}
		for _, pr := range v.Prices {
			pb := tx.VariantPrice.Create().
				SetVariantID(vEnt.ID).
				SetCurrencyCode(pr.CurrencyCode).
				SetAmount(pr.Amount)
			if pr.CompareAtAmount > 0 {
				pb.SetCompareAtAmount(pr.CompareAtAmount)
			}
			_, err = pb.Save(ctx)
			if err != nil {
				return nil, fmt.Errorf("create variant price: %w", err)
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return s.Get(ctx, p.ID)
}

func (s *Service) Get(ctx context.Context, id int) (*ent.Product, error) {
	p, err := s.entClient.Product.Query().
		Where(product.IDEQ(id)).
		WithTranslations().
		WithVariants(func(q *ent.VariantQuery) { q.WithPrices() }).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, pkgerr.ErrNotFound
		}
		return nil, fmt.Errorf("query product: %w", err)
	}
	return p, nil
}

func (s *Service) GetBySlug(ctx context.Context, slug string) (*ent.Product, error) {
	p, err := s.entClient.Product.Query().
		Where(product.SlugEQ(slug)).
		WithTranslations().
		WithVariants(func(q *ent.VariantQuery) { q.WithPrices() }).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, pkgerr.ErrNotFound
		}
		return nil, fmt.Errorf("query product: %w", err)
	}
	return p, nil
}

type ProductFilter struct {
	Status      string
	CategoryID  int
	ProductType string
	Page        int
	PageSize    int
}

func (s *Service) List(ctx context.Context, f ProductFilter) ([]*ent.Product, int, error) {
	q := s.entClient.Product.Query()
	if f.Status != "" {
		q = q.Where(product.StatusEQ(product.Status(f.Status)))
	}
	if f.CategoryID > 0 {
		q = q.Where(product.CategoryIDEQ(f.CategoryID))
	}
	if f.ProductType != "" {
		q = q.Where(product.ProductTypeEQ(product.ProductType(f.ProductType)))
	}
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 || f.PageSize > 100 {
		f.PageSize = 20
	}
	items, err := q.
		WithTranslations().
		WithVariants(func(q *ent.VariantQuery) { q.WithPrices() }).
		Order(ent.Desc(product.FieldID)).
		Limit(f.PageSize).
		Offset((f.Page - 1) * f.PageSize).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (s *Service) Publish(ctx context.Context, id int) error {
	return s.entClient.Product.UpdateOneID(id).
		SetStatus(product.StatusPublished).
		SetPublishedAt(time.Now()).
		Exec(ctx)
}

func (s *Service) Archive(ctx context.Context, id int) error {
	return s.entClient.Product.UpdateOneID(id).
		SetStatus(product.StatusArchived).
		Exec(ctx)
}

func (s *Service) Delete(ctx context.Context, id int) error {
	return s.entClient.Product.DeleteOneID(id).Exec(ctx)
}

func (s *Service) UpdateVariantPrice(ctx context.Context, variantID int, currencyCode string, amount, compareAtAmount int64) error {
	_, err := s.entClient.VariantPrice.Query().
		Where(variantprice.VariantIDEQ(variantID), variantprice.CurrencyCodeEQ(currencyCode)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			_, err = s.entClient.VariantPrice.Create().
				SetVariantID(variantID).
				SetCurrencyCode(currencyCode).
				SetAmount(amount).
				SetCompareAtAmount(compareAtAmount).
				Save(ctx)
			return err
		}
		return err
	}
	return s.entClient.VariantPrice.Update().
		Where(variantprice.VariantIDEQ(variantID), variantprice.CurrencyCodeEQ(currencyCode)).
		SetAmount(amount).
		SetCompareAtAmount(compareAtAmount).
		Exec(ctx)
}

func (s *Service) UpdateInventory(ctx context.Context, variantID int, qty int) error {
	v, err := s.entClient.Variant.Query().Where(variant.IDEQ(variantID)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return pkgerr.ErrNotFound
		}
		return err
	}
	return v.Update().SetInventory(qty).Exec(ctx)
}

func (s *Service) ListTranslations(ctx context.Context, productID int) ([]*ent.ProductTranslation, error) {
	return s.entClient.ProductTranslation.Query().
		Where(producttranslation.ProductIDEQ(productID)).
		All(ctx)
}

func (s *Service) Search(ctx context.Context, query, locale string, page, pageSize int) ([]*ent.Product, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	rows, err := s.db.QueryContext(ctx, `
		SELECT DISTINCT p.id
		FROM products p
		JOIN product_translations pt ON pt.product_id = p.id
		WHERE p.status = 'published'
		  AND pt.locale = $1
		  AND to_tsvector('simple', coalesce(pt.title, '') || ' ' || coalesce(pt.description, '')) @@ plainto_tsquery('simple', $2)
		ORDER BY p.id DESC
		LIMIT $3 OFFSET $4`,
		locale, query, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("search products: %w", err)
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, 0, err
		}
		ids = append(ids, id)
	}

	var total int
	_ = s.db.QueryRowContext(ctx, `
		SELECT count(DISTINCT p.id)
		FROM products p
		JOIN product_translations pt ON pt.product_id = p.id
		WHERE p.status = 'published'
		  AND pt.locale = $1
		  AND to_tsvector('simple', coalesce(pt.title, '') || ' ' || coalesce(pt.description, '')) @@ plainto_tsquery('simple', $2)`,
		locale, query).Scan(&total)

	if len(ids) == 0 {
		return []*ent.Product{}, 0, nil
	}
	items, err := s.entClient.Product.Query().
		Where(product.IDIn(ids...)).
		WithTranslations().
		WithVariants(func(q *ent.VariantQuery) { q.WithPrices() }).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}
