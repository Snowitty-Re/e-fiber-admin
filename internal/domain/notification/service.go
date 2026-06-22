package notification

import (
	"context"
	"fmt"
	"strings"

	"github.com/Snowitty-Re/e-fiber-admin/internal/ent"
	"github.com/Snowitty-Re/e-fiber-admin/internal/ent/emailtemplate"
	"github.com/Snowitty-Re/e-fiber-admin/internal/ent/notification"
)

type Service struct {
	entClient *ent.Client
}

func NewService(entClient *ent.Client) *Service {
	return &Service{entClient: entClient}
}

type RenderedEmail struct {
	Subject string
	BodyHTML string
	Locale  string
}

func (s *Service) Render(ctx context.Context, code, locale, recipient string, payload map[string]any) (*RenderedEmail, error) {
	tmpl, err := s.entClient.EmailTemplate.Query().
		Where(emailtemplate.CodeEQ(code)).
		WithTranslations().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("template %s not found", code)
		}
		return nil, fmt.Errorf("query template: %w", err)
	}

	var tr *ent.EmailTemplateTranslation
	for _, t := range tmpl.Edges.Translations {
		if t.Locale == locale {
			tr = t
			break
		}
	}
	if tr == nil {
		for _, t := range tmpl.Edges.Translations {
			tr = t
			break
		}
	}
	if tr == nil {
		return nil, fmt.Errorf("no template translation for %s (locale %s)", code, locale)
	}

	subject := renderVars(tr.Subject, payload)
	body := renderVars(tr.BodyHTML, payload)

	_, err = s.entClient.Notification.Create().
		SetChannel(notification.ChannelEmail).
		SetRecipient(recipient).
		SetTemplateCode(code).
		SetPayload(payload).
		SetStatus(notification.StatusSent).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("record notification: %w", err)
	}

	return &RenderedEmail{Subject: subject, BodyHTML: body, Locale: tr.Locale}, nil
}

func (s *Service) RecordFailed(ctx context.Context, recipient, code string, payload map[string]any, attempts int) error {
	return s.entClient.Notification.Create().
		SetChannel(notification.ChannelEmail).
		SetRecipient(recipient).
		SetTemplateCode(code).
		SetPayload(payload).
		SetStatus(notification.StatusFailed).
		SetAttempts(attempts).
		Exec(ctx)
}

func (s *Service) SeedTemplate(ctx context.Context, code string, variables []string, translations []TemplateTranslationInput) error {
	exists, err := s.entClient.EmailTemplate.Query().Where(emailtemplate.CodeEQ(code)).Count(ctx)
	if err != nil {
		return err
	}
	if exists > 0 {
		return nil
	}
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	t, err := tx.EmailTemplate.Create().
		SetCode(code).
		SetVariables(variables).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("create template: %w", err)
	}
	for _, ti := range translations {
		_, err = tx.EmailTemplateTranslation.Create().
			SetEmailTemplateID(t.ID).
			SetLocale(ti.Locale).
			SetSubject(ti.Subject).
			SetBodyHTML(ti.BodyHTML).
			Save(ctx)
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("create template translation: %w", err)
		}
	}
	return tx.Commit()
}

type TemplateTranslationInput struct {
	Locale   string
	Subject  string
	BodyHTML string
}

func renderVars(tmpl string, vars map[string]any) string {
	for k, v := range vars {
		placeholder := "{{" + k + "}}"
		tmpl = strings.ReplaceAll(tmpl, placeholder, fmt.Sprintf("%v", v))
	}
	return tmpl
}