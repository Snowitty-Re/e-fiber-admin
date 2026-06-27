package main

import (
	"context"
	"log/slog"

	"github.com/Snowitty-Re/e-fiber-admin/internal/domain/notification"
	"github.com/Snowitty-Re/e-fiber-admin/internal/ent"
)

func SeedNotifications(ctx context.Context, entClient *ent.Client) error {
	ns := notification.NewService(entClient)
	templates := []struct {
		Code      string
		Variables []string
		En        notification.TemplateTranslationInput
		Zh        notification.TemplateTranslationInput
	}{
		{
			Code:      "inquiry_received",
			Variables: []string{"name", "email", "form_slug", "inquiry_id"},
			En: notification.TemplateTranslationInput{
				Locale: "en", Subject: "We received your inquiry (#{{inquiry_id}})",
				BodyHTML: `<p>Hi {{name}},</p><p>We've received your inquiry on form <strong>{{form_slug}}</strong>.</p><p>Our team will get back to you shortly.</p>`,
			},
			Zh: notification.TemplateTranslationInput{
				Locale: "zh", Subject: "我们已收到您的询盘 (#{{inquiry_id}})",
				BodyHTML: `<p>您好 {{name}}，</p><p>我们已收到您在表单 <strong>{{form_slug}}</strong> 上的询盘，将有专人尽快与您联系。</p>`,
			},
		},
		{
			Code:      "inquiry_received_store",
			Variables: []string{"name", "email", "form_slug", "inquiry_id"},
			En: notification.TemplateTranslationInput{
				Locale: "en", Subject: "New inquiry #{{inquiry_id}} from {{email}}",
				BodyHTML: `<p>New inquiry on form <strong>{{form_slug}}</strong>:</p><ul><li>Name: {{name}}</li><li>Email: {{email}}</li><li>Inquiry ID: {{inquiry_id}}</li></ul>`,
			},
			Zh: notification.TemplateTranslationInput{
				Locale: "zh", Subject: "新询盘 #{{inquiry_id}} 来自 {{email}}",
				BodyHTML: `<p>新询盘（表单 <strong>{{form_slug}}</strong>）：</p><ul><li>姓名：{{name}}</li><li>邮箱：{{email}}</li><li>询盘 ID：{{inquiry_id}}</li></ul>`,
			},
		},
	}
	for _, t := range templates {
		translations := []notification.TemplateTranslationInput{t.En, t.Zh}
		if err := ns.SeedTemplate(ctx, t.Code, t.Variables, translations); err != nil {
			slog.Warn("seed template skipped", "code", t.Code, "err", err)
		}
	}
	slog.Info("notification templates seeded", "count", len(templates))
	return nil
}

func SeedOrderTemplates(ctx context.Context, entClient *ent.Client) error {
	ns := notification.NewService(entClient)
	templates := []struct {
		Code      string
		Variables []string
		En        notification.TemplateTranslationInput
	}{
		{
			Code:      "order_placed",
			Variables: []string{"order_number", "order_id", "total", "currency"},
			En: notification.TemplateTranslationInput{
				Locale: "en", Subject: "Order confirmation: {{order_number}}",
				BodyHTML: `<p>Thank you for your order <strong>{{order_number}}</strong>.</p><p>Total: {{currency}} {{total}}</p><p>We'll notify you when it ships.</p>`,
			},
		},
		{
			Code:      "order_cancelled",
			Variables: []string{"order_id"},
			En: notification.TemplateTranslationInput{
				Locale: "en", Subject: "Your order has been cancelled",
				BodyHTML: `<p>Your order (ID: {{order_id}}) has been cancelled.</p><p>Any payment will be refunded.</p>`,
			},
		},
	}
	seeded := 0
	for _, t := range templates {
		if err := ns.SeedTemplate(ctx, t.Code, t.Variables, []notification.TemplateTranslationInput{t.En}); err != nil {
			slog.Warn("seed order template skipped", "code", t.Code, "err", err)
			continue
		}
		seeded++
	}
	slog.Info("order templates seeded", "count", seeded)
	return nil
}
