package kafkafx

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type Config struct {
	Brokers          []string      `mapstructure:"brokers"`
	ClientID         string        `mapstructure:"client_id"`
	GroupID          string        `mapstructure:"group_id"`
	PublishTimeout   time.Duration `mapstructure:"publish_timeout"`
	SubscribeTimeout time.Duration `mapstructure:"subscribe_timeout"`
	MaxRetries       int           `mapstructure:"max_retries"`
	RetryInterval    time.Duration `mapstructure:"retry_interval"`
}

type KafkaClient struct {
	config   Config
	logger   *zap.Logger
	writers  map[string]*kafka.Writer
	readers  map[string]*kafka.Reader
	handlers map[string]MessageHandler
	mu       sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

func NewKafkaClient(config Config, logger *zap.Logger) (*KafkaClient, error) {
	if len(config.Brokers) == 0 {
		return nil, errors.New("no brokers provided")
	}

	if config.ClientID == "" {
		config.ClientID = "ecommerce"
	}

	if config.GroupID == "" {
		config.GroupID = "ecommerce-group"
	}

	if config.PublishTimeout == 0 {
		config.PublishTimeout = 5 * time.Second
	}

	if config.SubscribeTimeout == 0 {
		config.SubscribeTimeout = 5 * time.Second
	}

	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}

	if config.RetryInterval == 0 {
		config.RetryInterval = 1 * time.Second
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &KafkaClient{
		config:   config,
		logger:   logger,
		writers:  make(map[string]*kafka.Writer),
		readers:  make(map[string]*kafka.Reader),
		handlers: make(map[string]MessageHandler),
		ctx:      ctx,
		cancel:   cancel,
	}, nil
}

func (c *KafkaClient) Publish(ctx context.Context, topic string, message Message) error {
	if message.ID == "" {
		message.ID = generateMessageID()
	}
	if message.CreatedAt.IsZero() {
		message.CreatedAt = time.Now()
	}

	c.mu.RLock()
	writer, ok := c.writers[topic]
	c.mu.RUnlock()

	if !ok {
		var err error
		writer, err = c.createWriter(topic)
		if err != nil {
			return err
		}
	}

	value, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, c.config.PublishTimeout)
	defer cancel()

	err = writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(message.ID),
		Value: value,
		Time:  message.CreatedAt,
		Headers: []kafka.Header{
			{Key: "type", Value: []byte(message.Type)},
		},
	})
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return ErrPublishTimeout
		}
		return fmt.Errorf("failed to write message: %w", err)
	}

	c.logger.Debug("Message published",
		zap.String("topic", topic),
		zap.String("message_id", message.ID),
		zap.String("message_type", message.Type))

	return nil
}

func (c *KafkaClient) Subscribe(ctx context.Context, topic string, handler MessageHandler) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.handlers[topic]; ok {
		return fmt.Errorf("already subscribed to topic %s", topic)
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        c.config.Brokers,
		Topic:          topic,
		GroupID:        c.config.GroupID,
		MinBytes:       10e3, // 10KB
		MaxBytes:       10e6, // 10MB
		CommitInterval: time.Second,
		StartOffset:    kafka.FirstOffset,
	})

	c.readers[topic] = reader
	c.handlers[topic] = handler

	c.wg.Add(1)
	go c.consume(reader, topic, handler)

	c.logger.Info("Subscribed to topic", zap.String("topic", topic))

	return nil
}

func (c *KafkaClient) Unsubscribe(topic string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	reader, ok := c.readers[topic]
	if !ok {
		return ErrTopicNotFound
	}

	if err := reader.Close(); err != nil {
		return fmt.Errorf("failed to close reader: %w", err)
	}

	delete(c.handlers, topic)
	delete(c.readers, topic)

	c.logger.Info("Unsubscribed from topic", zap.String("topic", topic))

	return nil
}

func (c *KafkaClient) Close() error {
	c.cancel()

	c.mu.Lock()
	for topic, writer := range c.writers {
		if err := writer.Close(); err != nil {
			c.logger.Error("Failed to close writer",
				zap.String("topic", topic),
				zap.Error(err))
		}
	}

	for topic, reader := range c.readers {
		if err := reader.Close(); err != nil {
			c.logger.Error("Failed to close reader",
				zap.String("topic", topic),
				zap.Error(err))
		}
	}
	c.mu.Unlock()

	c.wg.Wait()

	c.logger.Info("Messaging client closed")

	return nil
}

func (c *KafkaClient) createWriter(topic string) (*kafka.Writer, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if writer, ok := c.writers[topic]; ok {
		return writer, nil
	}

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:      c.config.Brokers,
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    100,
		BatchTimeout: 100 * time.Millisecond,
		Async:        true,
	})

	c.writers[topic] = writer

	return writer, nil
}

func (c *KafkaClient) consume(reader *kafka.Reader, topic string, handler MessageHandler) {
	defer c.wg.Done()

	c.logger.Info("Starting consumer", zap.String("topic", topic))

	for {
		select {
		case <-c.ctx.Done():
			c.logger.Info("Consumer stopped", zap.String("topic", topic))
			return
		default:
			msg, err := reader.ReadMessage(c.ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return
				}
				c.logger.Error("Failed to read message",
					zap.String("topic", topic),
					zap.Error(err))
				time.Sleep(c.config.RetryInterval)
				continue
			}

			var message Message
			if err := json.Unmarshal(msg.Value, &message); err != nil {
				c.logger.Error("Failed to unmarshal message",
					zap.String("topic", topic),
					zap.Error(err))
				continue
			}

			handlerCtx, cancel := context.WithTimeout(c.ctx, c.config.SubscribeTimeout)

			if err := handler(handlerCtx, message); err != nil {
				c.logger.Error("Failed to handle message",
					zap.String("topic", topic),
					zap.String("message_id", message.ID),
					zap.Error(err))
			} else {
				c.logger.Debug("Message handled",
					zap.String("topic", topic),
					zap.String("message_id", message.ID),
					zap.String("message_type", message.Type))
			}

			cancel()
		}
	}
}

func generateMessageID() string {
	return fmt.Sprintf("msg-%d", time.Now().UnixNano())
}
