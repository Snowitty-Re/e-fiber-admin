package inquiry

import (
	"context"
	"fmt"

	"github.com/Snowitty-Re/e-fiber-admin/internal/ent"
	"github.com/Snowitty-Re/e-fiber-admin/internal/ent/formdefinition"
	"github.com/Snowitty-Re/e-fiber-admin/internal/ent/inquiry"
	"github.com/Snowitty-Re/e-fiber-admin/internal/events"
	pkgerr "github.com/Snowitty-Re/e-fiber-admin/internal/pkg/errors"
)

type Service struct {
	entClient *ent.Client
	bus       *events.Bus
}

func NewService(entClient *ent.Client, bus *events.Bus) *Service {
	return &Service{entClient: entClient, bus: bus}
}

type FormInput struct {
	Slug         string
	Fields       []map[string]any
	NotifyEmails []string
	IsActive     bool
	Translations []FormTranslationInput
}

type FormTranslationInput struct {
	Locale      string
	Title       string
	FieldLabels map[string]string
}

func (s *Service) CreateForm(ctx context.Context, in FormInput) (*ent.FormDefinition, error) {
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	f, err := tx.FormDefinition.Create().
		SetSlug(in.Slug).
		SetFields(in.Fields).
		SetNotifyEmails(in.NotifyEmails).
		SetIsActive(in.IsActive).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create form: %w", err)
	}
	for _, t := range in.Translations {
		_, err = tx.FormDefinitionTranslation.Create().
			SetFormDefinitionID(f.ID).
			SetLocale(t.Locale).
			SetTitle(t.Title).
			SetFieldLabels(t.FieldLabels).
			Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("create form translation: %w", err)
		}
	}
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return s.GetForm(ctx, f.ID)
}

func (s *Service) GetForm(ctx context.Context, id int) (*ent.FormDefinition, error) {
	f, err := s.entClient.FormDefinition.Query().
		Where(formdefinition.IDEQ(id)).
		WithTranslations().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, pkgerr.ErrNotFound
		}
		return nil, fmt.Errorf("query form: %w", err)
	}
	return f, nil
}

func (s *Service) GetFormBySlug(ctx context.Context, slug string) (*ent.FormDefinition, error) {
	f, err := s.entClient.FormDefinition.Query().
		Where(formdefinition.SlugEQ(slug), formdefinition.IsActiveEQ(true)).
		WithTranslations().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, pkgerr.ErrNotFound
		}
		return nil, fmt.Errorf("query form: %w", err)
	}
	return f, nil
}

func (s *Service) ListForms(ctx context.Context) ([]*ent.FormDefinition, error) {
	return s.entClient.FormDefinition.Query().WithTranslations().All(ctx)
}

type InquirySubmitInput struct {
	FormSlug   string
	CustomerID int
	Email      string
	Phone      string
	Name       string
	Company    string
	Payload    map[string]any
	ProductID  int
}

func (s *Service) SubmitInquiry(ctx context.Context, in InquirySubmitInput) (*ent.Inquiry, error) {
	form, err := s.entClient.FormDefinition.Query().
		Where(formdefinition.SlugEQ(in.FormSlug), formdefinition.IsActiveEQ(true)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, pkgerr.New("FORM_NOT_FOUND", 404, "form not found or inactive")
		}
		return nil, fmt.Errorf("query form: %w", err)
	}

	b := s.entClient.Inquiry.Create().
		SetFormID(form.ID).
		SetEmail(in.Email).
		SetPhone(in.Phone).
		SetName(in.Name).
		SetCompany(in.Company).
		SetPayload(in.Payload).
		SetStatus(inquiry.StatusNew)
	if in.CustomerID > 0 {
		b.SetCustomerID(in.CustomerID)
	}
	if in.ProductID > 0 {
		b.SetProductID(in.ProductID)
	}

	inq, err := b.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create inquiry: %w", err)
	}

	if s.bus != nil {
		_ = s.bus.PublishSimple(ctx, "inquiry.received", "inquiry", fmt.Sprintf("%d", inq.ID), map[string]any{
			"inquiry_id": inq.ID, "email": inq.Email, "name": inq.Name,
			"form_slug": in.FormSlug, "form_id": form.ID,
		})
	}
	return inq, nil
}

func (s *Service) ListInquiries(ctx context.Context, page, pageSize int, status string) ([]*ent.Inquiry, int, error) {
	q := s.entClient.Inquiry.Query()
	if status != "" {
		q = q.Where(inquiry.StatusEQ(inquiry.Status(status)))
	}
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	items, err := q.Order(ent.Desc(inquiry.FieldID)).Limit(pageSize).Offset((page - 1) * pageSize).All(ctx)
	return items, total, err
}

func (s *Service) GetInquiry(ctx context.Context, id int) (*ent.Inquiry, error) {
	inq, err := s.entClient.Inquiry.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, pkgerr.ErrNotFound
		}
		return nil, fmt.Errorf("query inquiry: %w", err)
	}
	return inq, nil
}

func (s *Service) AssignInquiry(ctx context.Context, id, adminID int) error {
	if err := s.entClient.Inquiry.UpdateOneID(id).
		SetAssignedAdminID(adminID).
		SetStatus(inquiry.StatusContacted).
		Exec(ctx); err != nil {
		return err
	}
	if s.bus != nil {
		_ = s.bus.PublishSimple(ctx, "inquiry.assigned", "inquiry", fmt.Sprintf("%d", id), map[string]any{
			"inquiry_id": id, "admin_id": adminID,
		})
	}
	return nil
}

func (s *Service) UpdateStatus(ctx context.Context, id int, status string) error {
	if err := s.entClient.Inquiry.UpdateOneID(id).
		SetStatus(inquiry.Status(status)).
		Exec(ctx); err != nil {
		return err
	}
	if s.bus != nil {
		_ = s.bus.PublishSimple(ctx, "inquiry.updated", "inquiry", fmt.Sprintf("%d", id), map[string]any{
			"inquiry_id": id, "status": status,
		})
	}
	return nil
}

func (s *Service) ConvertToOrder(ctx context.Context, inquiryID, orderID int) error {
	if err := s.entClient.Inquiry.UpdateOneID(inquiryID).
		SetConvertedOrderID(orderID).
		SetStatus(inquiry.StatusConverted).
		Exec(ctx); err != nil {
		return err
	}
	if s.bus != nil {
		_ = s.bus.PublishSimple(ctx, "inquiry.converted", "inquiry", fmt.Sprintf("%d", inquiryID), map[string]any{
			"inquiry_id": inquiryID, "order_id": orderID,
		})
	}
	return nil
}