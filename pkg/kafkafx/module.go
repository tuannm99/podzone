package kafkafx

import (
	"context"
	"errors"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module provides Kafka client
var Module = fx.Options(
	fx.Provide(
		func(logger *zap.Logger) (*KafkaClient, error) {
			config := &Config{
				Brokers: []string{"localhost:9092"},
				GroupID: "ecommerce-group",
			}
			return NewKafkaClient(config, logger)
		},
	),
	fx.Invoke(RegisterLifecycle),
)

func RegisterLifecycle(lc fx.Lifecycle, logger *zap.Logger) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			logger.Info("Closing Redis connection")
			return nil
		},
	})
}

var (
	ErrPublishTimeout   = errors.New("publish timeout")
	ErrSubscribeTimeout = errors.New("subscribe timeout")
	ErrConnectionClosed = errors.New("connection closed")
	ErrInvalidMessage   = errors.New("invalid message")
	ErrTopicNotFound    = errors.New("topic not found")
	ErrHandlerNotFound  = errors.New("handler not found")
)

type Message struct {
	ID        string            `json:"id"`
	Type      string            `json:"type"`
	Data      map[string]any    `json:"data"`
	Metadata  map[string]string `json:"metadata"`
	CreatedAt time.Time         `json:"created_at"`
}

type Publisher interface {
	Publish(ctx context.Context, topic string, message Message) error
	Close() error
}

type Subscriber interface {
	Subscribe(ctx context.Context, topic string, handler MessageHandler) error
	Unsubscribe(topic string) error
	Close() error
}

type MessageHandler func(ctx context.Context, message Message) error

type Client interface {
	Publisher
	Subscriber
}
