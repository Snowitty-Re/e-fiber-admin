package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Event struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	OccurredAt  time.Time `json:"occurred_at"`
	Aggregate   string    `json:"aggregate"`
	AggregateID string    `json:"aggregate_id"`
	ActorType   string    `json:"actor_type,omitempty"`
	ActorID     string    `json:"actor_id,omitempty"`
	Data        any       `json:"data,omitempty"`
	Version     int       `json:"version"`
}

type Handler func(ctx context.Context, event *Event) error

type Bus struct {
	redisClient *redis.Client
}

func NewBus(redisClient *redis.Client) *Bus {
	return &Bus{redisClient: redisClient}
}

func channelName(eventName string) string {
	return "efa:events:" + eventName
}

func (b *Bus) Publish(ctx context.Context, event *Event) error {
	if event.ID == "" {
		event.ID = "evt_" + uuid.NewString()
	}
	if event.OccurredAt.IsZero() {
		event.OccurredAt = time.Now().UTC()
	}
	if event.Version == 0 {
		event.Version = 1
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}
	if err := b.redisClient.Publish(ctx, channelName(event.Name), payload).Err(); err != nil {
		return fmt.Errorf("publish event: %w", err)
	}
	return nil
}

func (b *Bus) PublishSimple(ctx context.Context, name, aggregate, aggregateID string, data any) error {
	return b.Publish(ctx, &Event{
		Name: name, Aggregate: aggregate, AggregateID: aggregateID, Data: data,
	})
}

type Subscriber struct {
	redisClient *redis.Client
	handlers    map[string][]Handler
}

func NewSubscriber(redisClient *redis.Client) *Subscriber {
	return &Subscriber{
		redisClient: redisClient,
		handlers:    make(map[string][]Handler),
	}
}

func (s *Subscriber) Subscribe(eventNames []string, handlers ...Handler) {
	for _, name := range eventNames {
		s.handlers[name] = append(s.handlers[name], handlers...)
	}
}

func (s *Subscriber) Run(ctx context.Context) error {
	if len(s.handlers) == 0 {
		slog.Info("no event subscribers registered, worker idle")
		return nil
	}
	var channels []string
	for name := range s.handlers {
		channels = append(channels, channelName(name))
	}

	pubsub := s.redisClient.Subscribe(ctx, channels...)
	defer pubsub.Close()

	ch := pubsub.Channel()
	slog.Info("event subscriber started", "channels", channels)

	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				return nil
			}
			var event Event
			if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
				slog.Error("unmarshal event failed", "err", err, "channel", msg.Channel)
				continue
			}
			s.dispatch(ctx, &event)
		case <-ctx.Done():
			slog.Info("event subscriber shutting down")
			return ctx.Err()
		}
	}
}

func (s *Subscriber) dispatch(ctx context.Context, event *Event) {
	handlers, ok := s.handlers[event.Name]
	if !ok {
		return
	}
	dedupKey := "efa:dedup:" + event.ID
	set, err := s.redisClient.SetNX(ctx, dedupKey, "1", 24*time.Hour).Result()
	if err != nil {
		slog.Error("dedup setnx failed", "err", err, "event_id", event.ID)
	}
	if !set {
		slog.Info("event already processed, skipping", "event_id", event.ID, "name", event.Name)
		return
	}
	for _, h := range handlers {
		if err := h(ctx, event); err != nil {
			slog.Error("event handler failed", "err", err, "event_id", event.ID, "name", event.Name)
		}
	}
}
