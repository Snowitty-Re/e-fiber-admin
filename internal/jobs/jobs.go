package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/hibiken/asynq"

	"github.com/Snowitty-Re/e-fiber-admin/internal/config"
)

const (
	QueueCritical = "critical"
	QueueDefault  = "default"
	QueueLow      = "low"

	TaskSendEmail     = "notification:send_email"
	TaskNotifyInquiry = "notification:notify_inquiry"
)

type Client struct {
	asynqClient *asynq.Client
}

func NewClient(cfg config.RedisConfig) *Client {
	return &Client{
		asynqClient: asynq.NewClient(asynq.RedisClientOpt{
			Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
			Password: cfg.Password,
			DB:       cfg.DB,
		}),
	}
}

func (c *Client) Enqueue(ctx context.Context, taskType string, payload []byte, queue string) error {
	task := asynq.NewTask(taskType, payload)
	opts := []asynq.Option{asynq.Queue(queue)}
	if queue == QueueCritical {
		opts = append(opts, asynq.MaxRetry(25))
	} else {
		opts = append(opts, asynq.MaxRetry(10))
	}
	info, err := c.asynqClient.EnqueueContext(ctx, task, opts...)
	if err != nil {
		return fmt.Errorf("enqueue %s: %w", taskType, err)
	}
	slog.Info("task enqueued", "type", taskType, "id", info.ID, "queue", info.Queue)
	return nil
}

type Server struct {
	asynqServer *asynq.Server
	mux         *asynq.ServeMux
}

func NewServer(cfg config.AsynqConfig, redisCfg config.RedisConfig) *Server {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     fmt.Sprintf("%s:%s", redisCfg.Host, redisCfg.Port),
			Password: redisCfg.Password,
			DB:       redisCfg.DB,
		},
		asynq.Config{
			Concurrency: cfg.Concurrency,
			Queues: map[string]int{
				QueueCritical: 6,
				QueueDefault:  3,
				QueueLow:      1,
			},
		},
	)
	return &Server{
		asynqServer: srv,
		mux:         asynq.NewServeMux(),
	}
}

func (s *Server) Handle(taskType string, handler func(context.Context, []byte) error) {
	s.mux.HandleFunc(taskType, func(ctx context.Context, t *asynq.Task) error {
		return handler(ctx, t.Payload())
	})
}

func (s *Server) Run() error {
	slog.Info("asynq worker starting")
	return s.asynqServer.Run(s.mux)
}

func (s *Server) Shutdown() {
	s.asynqServer.Shutdown()
}

func MustMarshal(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		slog.Error("marshal payload failed", "err", err)
		return nil
	}
	return b
}
