package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/Snowitty-Re/e-fiber-admin/internal/domain/notification"
	"github.com/Snowitty-Re/e-fiber-admin/internal/ent"
	"github.com/Snowitty-Re/e-fiber-admin/internal/events"
	"github.com/Snowitty-Re/e-fiber-admin/internal/jobs"
)

type InquiryReceivedPayload struct {
	InquiryID int    `json:"inquiry_id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	FormSlug  string `json:"form_slug"`
	FormID    int    `json:"form_id"`
}

type SendEmailPayload struct {
	TemplateCode string         `json:"template_code"`
	Locale       string         `json:"locale"`
	Recipient    string         `json:"recipient"`
	Vars         map[string]any `json:"vars"`
}

func RegisterEventHandlers(sub *events.Subscriber, entClient *ent.Client, jobsClient *jobs.Client) {
	sub.Subscribe([]string{"inquiry.received"},
		handleInquiryReceived(entClient, jobsClient),
	)
	sub.Subscribe([]string{"customer.registered"},
		handleCustomerRegistered(jobsClient),
	)
	sub.Subscribe([]string{"order.placed"},
		handleOrderPlaced(jobsClient),
	)
	sub.Subscribe([]string{"order.cancelled"},
		handleOrderCancelled(jobsClient),
	)
	sub.Subscribe([]string{"order.paid"},
		handleOrderPaid(jobsClient),
	)
	sub.Subscribe([]string{"order.fulfilled"},
		handleOrderFulfilled(jobsClient),
	)
	slog.Info("event handlers registered", "events", []string{
		"inquiry.received", "customer.registered",
		"order.placed", "order.cancelled", "order.paid", "order.fulfilled",
	})
}

func RegisterJobHandlers(s *jobs.Server, entClient *ent.Client) {
	s.Handle(jobs.TaskSendEmail, handleSendEmail(entClient))
	slog.Info("job handlers registered", "tasks", []string{jobs.TaskSendEmail})
}

func handleInquiryReceived(entClient *ent.Client, jobsClient *jobs.Client) events.Handler {
	return func(ctx context.Context, event *events.Event) error {
		var p InquiryReceivedPayload
		b, _ := json.Marshal(event.Data)
		if err := json.Unmarshal(b, &p); err != nil {
			return fmt.Errorf("unmarshal inquiry.received: %w", err)
		}
		slog.Info("inquiry.received event", "inquiry_id", p.InquiryID, "email", p.Email)

		form, err := entClient.FormDefinition.Get(ctx, p.FormID)
		if err != nil {
			return fmt.Errorf("query form: %w", err)
		}
		vars := map[string]any{
			"name":       p.Name,
			"email":      p.Email,
			"form_slug":  p.FormSlug,
			"inquiry_id": p.InquiryID,
		}
		customerPayload := jobs.MustMarshal(SendEmailPayload{
			TemplateCode: "inquiry_received",
			Locale:       "en",
			Recipient:    p.Email,
			Vars:         vars,
		})
		if err := jobsClient.Enqueue(ctx, jobs.TaskSendEmail, customerPayload, jobs.QueueDefault); err != nil {
			slog.Error("enqueue customer email failed", "err", err)
		}
		for _, e := range form.NotifyEmails {
			storePayload := jobs.MustMarshal(SendEmailPayload{
				TemplateCode: "inquiry_received_store",
				Locale:       "en",
				Recipient:    e,
				Vars:         vars,
			})
			if err := jobsClient.Enqueue(ctx, jobs.TaskSendEmail, storePayload, jobs.QueueCritical); err != nil {
				slog.Error("enqueue store email failed", "err", err, "recipient", e)
			}
		}
		return nil
	}
}

func handleCustomerRegistered(jobsClient *jobs.Client) events.Handler {
	return func(ctx context.Context, event *events.Event) error {
		slog.Info("customer.registered event", "data", event.Data)
		return nil
	}
}

func handleOrderPlaced(jobsClient *jobs.Client) events.Handler {
	return func(ctx context.Context, event *events.Event) error {
		slog.Info("order.placed event", "data", event.Data)
		p, ok := event.Data.(map[string]any)
		if !ok {
			return fmt.Errorf("invalid event data")
		}
		vars := map[string]any{
			"order_number": p["order_number"],
			"order_id":     p["order_id"],
			"total":        p["total"],
			"currency":     p["currency"],
		}
		payload := jobs.MustMarshal(SendEmailPayload{
			TemplateCode: "order_placed",
			Locale:       "en",
			Recipient:    fmt.Sprintf("%v", p["email"]),
			Vars:         vars,
		})
		return jobsClient.Enqueue(ctx, jobs.TaskSendEmail, payload, jobs.QueueCritical)
	}
}

func handleOrderCancelled(jobsClient *jobs.Client) events.Handler {
	return func(ctx context.Context, event *events.Event) error {
		slog.Info("order.cancelled event", "data", event.Data)
		p, ok := event.Data.(map[string]any)
		if !ok {
			return fmt.Errorf("invalid event data")
		}
		vars := map[string]any{"order_id": p["order_id"]}
		payload := jobs.MustMarshal(SendEmailPayload{
			TemplateCode: "order_cancelled",
			Locale:       "en",
			Recipient:    fmt.Sprintf("%v", p["email"]),
			Vars:         vars,
		})
		return jobsClient.Enqueue(ctx, jobs.TaskSendEmail, payload, jobs.QueueDefault)
	}
}

func handleOrderPaid(jobsClient *jobs.Client) events.Handler {
	return func(ctx context.Context, event *events.Event) error {
		slog.Info("order.paid event", "data", event.Data)
		return nil
	}
}

func handleOrderFulfilled(jobsClient *jobs.Client) events.Handler {
	return func(ctx context.Context, event *events.Event) error {
		slog.Info("order.fulfilled event", "data", event.Data)
		return nil
	}
}

func handleSendEmail(entClient *ent.Client) func(context.Context, []byte) error {
	notifService := notification.NewService(entClient)
	return func(ctx context.Context, payload []byte) error {
		var p SendEmailPayload
		if err := json.Unmarshal(payload, &p); err != nil {
			return fmt.Errorf("unmarshal send_email payload: %w", err)
		}
		rendered, err := notifService.Render(ctx, p.TemplateCode, p.Locale, p.Recipient, p.Vars)
		if err != nil {
			slog.Error("render email failed", "err", err, "template", p.TemplateCode)
			return err
		}
		slog.Info("email rendered & notification recorded",
			"template", p.TemplateCode,
			"recipient", p.Recipient,
			"subject", rendered.Subject,
			"locale", rendered.Locale,
			"body_len", len(rendered.BodyHTML),
		)
		return nil
	}
}
