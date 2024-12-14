package messaging

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

// Common errors
var (
	ErrPublishTimeout   = errors.New("publish timeout")
	ErrSubscribeTimeout = errors.New("subscribe timeout")
	ErrConnectionClosed = errors.New("connection closed")
	ErrInvalidMessage   = errors.New("invalid message")
	ErrTopicNotFound    = errors.New("topic not found")
	ErrHandlerNotFound  = errors.New("handler not found")
)

// Message represents a message
type Message struct {
	ID        string            `json:"id"`
	Type      string            `json:"type"`
	Data      map[string]any    `json:"data"`
	Metadata  map[string]string `json:"metadata"`
	CreatedAt time.Time         `json:"created_at"`
}

// Publisher defines methods for publishing messages
type Publisher interface {
	// Publish publishes a message to a topic
	Publish(ctx context.Context, topic string, message Message) error
	// Close closes the publisher
	Close() error
}

// Subscriber defines methods for subscribing to messages
type Subscriber interface {
	// Subscribe subscribes to a topic
	Subscribe(ctx context.Context, topic string, handler MessageHandler) error
	// Unsubscribe unsubscribes from a topic
	Unsubscribe(topic string) error
	// Close closes the subscriber
	Close() error
}

// MessageHandler is a function that handles messages
type MessageHandler func(ctx context.Context, message Message) error

// Client combines Publisher and Subscriber interfaces
type Client interface {
	Publisher
	Subscriber
}

// Config holds messaging client configuration
type Config struct {
	Brokers          []string      `mapstructure:"brokers"`
	ClientID         string        `mapstructure:"client_id"`
	GroupID          string        `mapstructure:"group_id"`
	PublishTimeout   time.Duration `mapstructure:"publish_timeout"`
	SubscribeTimeout time.Duration `mapstructure:"subscribe_timeout"`
	MaxRetries       int           `mapstructure:"max_retries"`
	RetryInterval    time.Duration `mapstructure:"retry_interval"`
}

// KafkaClient implements Client using Kafka
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

// NewKafkaClient creates a new Kafka client
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

// Publish publishes a message to a topic
func (c *KafkaClient) Publish(ctx context.Context, topic string, message Message) error {
	// Set message ID and timestamp if not set
	if message.ID == "" {
		message.ID = generateMessageID()
	}
	if message.CreatedAt.IsZero() {
		message.CreatedAt = time.Now()
	}

	// Get writer for topic
	c.mu.RLock()
	writer, ok := c.writers[topic]
	c.mu.RUnlock()

	// Create writer if not exists
	if !ok {
		var err error
		writer, err = c.createWriter(topic)
		if err != nil {
			return err
		}
	}

	// Marshal message
	value, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, c.config.PublishTimeout)
	defer cancel()

	// Write message
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

// Subscribe subscribes to a topic
func (c *KafkaClient) Subscribe(ctx context.Context, topic string, handler MessageHandler) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if already subscribed
	if _, ok := c.handlers[topic]; ok {
		return fmt.Errorf("already subscribed to topic %s", topic)
	}

	// Create reader
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        c.config.Brokers,
		Topic:          topic,
		GroupID:        c.config.GroupID,
		MinBytes:       10e3, // 10KB
		MaxBytes:       10e6, // 10MB
		CommitInterval: time.Second,
		StartOffset:    kafka.FirstOffset,
	})

	// Store reader and handler
	c.readers[topic] = reader
	c.handlers[topic] = handler

	// Start consumer goroutine
	c.wg.Add(1)
	go c.consume(reader, topic, handler)

	c.logger.Info("Subscribed to topic", zap.String("topic", topic))

	return nil
}

// Unsubscribe unsubscribes from a topic
func (c *KafkaClient) Unsubscribe(topic string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if subscribed
	reader, ok := c.readers[topic]
	if !ok {
		return ErrTopicNotFound
	}

	// Close reader
	if err := reader.Close(); err != nil {
		return fmt.Errorf("failed to close reader: %w", err)
	}

	// Remove handler and reader
	delete(c.handlers, topic)
	delete(c.readers, topic)

	c.logger.Info("Unsubscribed from topic", zap.String("topic", topic))

	return nil
}

// Close closes the client
func (c *KafkaClient) Close() error {
	c.cancel()

	// Close all writers
	c.mu.Lock()
	for topic, writer := range c.writers {
		if err := writer.Close(); err != nil {
			c.logger.Error("Failed to close writer",
				zap.String("topic", topic),
				zap.Error(err))
		}
	}

	// Close all readers
	for topic, reader := range c.readers {
		if err := reader.Close(); err != nil {
			c.logger.Error("Failed to close reader",
				zap.String("topic", topic),
				zap.Error(err))
		}
	}
	c.mu.Unlock()

	// Wait for all consumers to exit
	c.wg.Wait()

	c.logger.Info("Messaging client closed")

	return nil
}

// createWriter creates a writer for a topic
func (c *KafkaClient) createWriter(topic string) (*kafka.Writer, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if writer already exists
	if writer, ok := c.writers[topic]; ok {
		return writer, nil
	}

	// Create writer
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:      c.config.Brokers,
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    100,
		BatchTimeout: 100 * time.Millisecond,
		Async:        true,
	})

	// Store writer
	c.writers[topic] = writer

	return writer, nil
}

// consume consumes messages from a topic
func (c *KafkaClient) consume(reader *kafka.Reader, topic string, handler MessageHandler) {
	defer c.wg.Done()

	c.logger.Info("Starting consumer", zap.String("topic", topic))

	for {
		select {
		case <-c.ctx.Done():
			c.logger.Info("Consumer stopped", zap.String("topic", topic))
			return
		default:
			// Read message
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

			// Parse message
			var message Message
			if err := json.Unmarshal(msg.Value, &message); err != nil {
				c.logger.Error("Failed to unmarshal message",
					zap.String("topic", topic),
					zap.Error(err))
				continue
			}

			// Create context for handler
			handlerCtx, cancel := context.WithTimeout(c.ctx, c.config.SubscribeTimeout)

			// Handle message
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

// generateMessageID generates a unique message ID
func generateMessageID() string {
	return fmt.Sprintf("msg-%d", time.Now().UnixNano())
}

// Event types
const (
	EventTypeOrderCreated     = "order.created"
	EventTypeOrderUpdated     = "order.updated"
	EventTypeOrderCancelled   = "order.cancelled"
	EventTypeOrderCompleted   = "order.completed"
	EventTypePaymentProcessed = "payment.processed"
	EventTypePaymentFailed    = "payment.failed"
	EventTypeProductCreated   = "product.created"
	EventTypeProductUpdated   = "product.updated"
	EventTypeProductDeleted   = "product.deleted"
	EventTypeUserRegistered   = "user.registered"
	EventTypeUserUpdated      = "user.updated"
)

// Topic names
const (
	TopicOrders    = "orders"
	TopicPayments  = "payments"
	TopicProducts  = "products"
	TopicUsers     = "users"
	TopicInventory = "inventory"
)

// NewLocalClient creates a new in-memory client for testing
type LocalClient struct {
	handlers map[string][]MessageHandler
	mu       sync.RWMutex
	logger   *zap.Logger
}

// NewLocalClient creates a new local client
func NewLocalClient(logger *zap.Logger) *LocalClient {
	return &LocalClient{
		handlers: make(map[string][]MessageHandler),
		logger:   logger,
	}
}

// Publish publishes a message locally
func (c *LocalClient) Publish(ctx context.Context, topic string, message Message) error {
	// Set message ID and timestamp if not set
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

	// Call all handlers
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

// Subscribe subscribes to a topic locally
func (c *LocalClient) Subscribe(ctx context.Context, topic string, handler MessageHandler) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Add handler
	if _, ok := c.handlers[topic]; !ok {
		c.handlers[topic] = []MessageHandler{}
	}
	c.handlers[topic] = append(c.handlers[topic], handler)

	c.logger.Info("Subscribed to topic locally", zap.String("topic", topic))

	return nil
}

// Unsubscribe unsubscribes from a topic locally
func (c *LocalClient) Unsubscribe(topic string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Remove all handlers for topic
	delete(c.handlers, topic)

	c.logger.Info("Unsubscribed from topic locally", zap.String("topic", topic))

	return nil
}

// Close closes the local client
func (c *LocalClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Clear all handlers
	c.handlers = make(map[string][]MessageHandler)

	c.logger.Info("Local messaging client closed")

	return nil
}
