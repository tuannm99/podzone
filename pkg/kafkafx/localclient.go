package kafkafx

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

type LocalClient struct {
	handlers map[string][]MessageHandler
	mu       sync.RWMutex
	logger   *zap.Logger
}

func NewLocalClient(logger *zap.Logger) *LocalClient {
	return &LocalClient{
		handlers: make(map[string][]MessageHandler),
		logger:   logger,
	}
}

func (c *LocalClient) Publish(ctx context.Context, topic string, message Message) error {
	if message.ID == "" {
		message.ID = generateMessageID()
	}
	if message.CreatedAt.IsZero() {
		message.CreatedAt = time.Now()
	}

	c.mu.RLock()
	handlers, ok := c.handlers[topic]
	c.mu.RUnlock()

	if !ok || len(handlers) == 0 {
		c.logger.Warn("No handlers for topic", zap.String("topic", topic))
		return nil
	}

	for _, handler := range handlers {
		if err := handler(ctx, message); err != nil {
			c.logger.Error("Failed to handle message",
				zap.String("topic", topic),
				zap.String("message_id", message.ID),
				zap.Error(err))
		}
	}

	c.logger.Debug("Message published locally",
		zap.String("topic", topic),
		zap.String("message_id", message.ID),
		zap.String("message_type", message.Type))

	return nil
}

func (c *LocalClient) Subscribe(ctx context.Context, topic string, handler MessageHandler) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.handlers[topic]; !ok {
		c.handlers[topic] = []MessageHandler{}
	}
	c.handlers[topic] = append(c.handlers[topic], handler)

	c.logger.Info("Subscribed to topic locally", zap.String("topic", topic))

	return nil
}

func (c *LocalClient) Unsubscribe(topic string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.handlers, topic)

	c.logger.Info("Unsubscribed from topic locally", zap.String("topic", topic))

	return nil
}

func (c *LocalClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.handlers = make(map[string][]MessageHandler)

	c.logger.Info("Local messaging client closed")

	return nil
}

func generateMessageID() string {
	return fmt.Sprintf("msg-%d", time.Now().UnixNano())
}
