package main

import (
	"log/slog"

	"github.com/redis/go-redis/v9"

	"github.com/Snowitty-Re/e-fiber-admin/internal/config"
	"github.com/Snowitty-Re/e-fiber-admin/internal/events"
	"github.com/Snowitty-Re/e-fiber-admin/internal/jobs"
)

func RegisterEventHandlers(sub *events.Subscriber, cfg *config.Config, rc *redis.Client) []string {
	var registered []string
	for name := range registeredEventNames() {
		registered = append(registered, name)
	}
	slog.Info("event handlers registered", "count", len(registered))
	return registered
}

func RegisterJobHandlers(s *jobs.Server, cfg *config.Config, rc *redis.Client) {
	slog.Info("job handlers registered")
}

func registeredEventNames() map[string]bool {
	return map[string]bool{
		"inquiry.received":    true,
		"inquiry.updated":     true,
		"inquiry.assigned":    true,
		"inquiry.converted":   true,
		"customer.registered": true,
	}
}