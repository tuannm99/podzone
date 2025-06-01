package kafkafx

import (
	"context"
	"errors"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Options(
	fx.Provide(
		func(logger *zap.Logger) (*KafkaClient, error) {
			config := &Config{
				Brokers: []string{"localhost:9092"},
				GroupID: "ecommerce-group",
			}
			return NewKafkaClient(config, logger)
		},
		func(client *KafkaClient, logger *zap.Logger) (*PubSubClient, error) {
			configs := DefaultConfigs(client.config.Brokers)
			pubsubConfig := configs["pubsub"].(PubSubConfig)
			return NewPubSubClient(client, &pubsubConfig, logger)
		},
		func(client *KafkaClient, logger *zap.Logger) (*QueueClient, error) {
			configs := DefaultConfigs(client.config.Brokers)
			queueConfig := configs["queue"].(QueueConfig)
			return NewQueueClient(client, &queueConfig, logger)
		},
		func(client *KafkaClient, logger *zap.Logger) (*ConsumerGroupClient, error) {
			configs := DefaultConfigs(client.config.Brokers)
			groupConfig := configs["consumer_group"].(ConsumerGroupConfig)
			return NewConsumerGroupClient(&groupConfig, logger)
		},
		func(client *KafkaClient, logger *zap.Logger) (*ExactlyOnceClient, error) {
			configs := DefaultConfigs(client.config.Brokers)
			exactConfig := configs["exactly_once"].(ExactlyOnceConfig)
			return NewExactlyOnceClient(client, &exactConfig, logger)
		},
	),
	fx.Invoke(RegisterLifecycle),
)

func RegisterLifecycle(lc fx.Lifecycle, logger *zap.Logger) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			logger.Info("Closing Kafka connection")
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
