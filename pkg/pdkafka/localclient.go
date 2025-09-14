package pdkafka

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/tuannm99/podzone/pkg/pdlog"
)

type LocalClient struct {
	handlers map[string][]MessageHandler
	mu       sync.RWMutex
	logger   pdlog.Logger
}

func NewLocalClient(logger pdlog.Logger) *LocalClient {
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
		c.logger.Warn("No handlers for topic").With("topic", topic).Send()
		return nil
	}

	for _, handler := range handlers {
		if err := handler(ctx, message); err != nil {
			c.logger.Error("Failed to handle message").
				With("topic", topic).
				With("message_id", message.ID).
				Err(err).
				Send()
		}
	}

	c.logger.Debug("Message published locally").
		With("topic", topic).
		With("message_id", message.ID).
		With("message_type", message.Type).
		Send()

	return nil
}

func (c *LocalClient) Subscribe(ctx context.Context, topic string, handler MessageHandler) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.handlers[topic]; !ok {
		c.handlers[topic] = []MessageHandler{}
	}
	c.handlers[topic] = append(c.handlers[topic], handler)

	c.logger.Info("Subscribed to topic locally").With("topic", topic).Send()

	return nil
}

func (c *LocalClient) Unsubscribe(topic string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.handlers, topic)

	c.logger.Info("Unsubscribed from topic locally").With("topic", topic).Send()

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
