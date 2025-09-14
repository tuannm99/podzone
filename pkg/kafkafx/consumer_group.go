package kafkafx

import (
	"context"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/tuannm99/podzone/pkg/pdlog"
)

// ConsumerGroupClient implements consumer group pattern
type ConsumerGroupClient struct {
	logger pdlog.Logger
	config *ConsumerGroupConfig
	group  sarama.ConsumerGroup
}

func NewConsumerGroupClient(config *ConsumerGroupConfig, logger pdlog.Logger) (*ConsumerGroupClient, error) {
	saramaConfig := config.ToSaramaConfig()

	group, err := sarama.NewConsumerGroup(config.Brokers, config.GroupID, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer group: %w", err)
	}

	return &ConsumerGroupClient{
		logger: logger,
		config: config,
		group:  group,
	}, nil
}

func (c *ConsumerGroupClient) Consume(ctx context.Context, topics []string, handler func([]byte) error) error {
	consumer := &consumerGroupHandler{
		logger:  c.logger,
		handler: handler,
	}

	for {
		err := c.group.Consume(ctx, topics, consumer)
		if err != nil {
			return fmt.Errorf("failed to consume: %w", err)
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
}

func (c *ConsumerGroupClient) Close() error {
	return c.group.Close()
}

type consumerGroupHandler struct {
	logger  pdlog.Logger
	handler func([]byte) error
}

func (h *consumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *consumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (h *consumerGroupHandler) ConsumeClaim(
	session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim,
) error {
	for msg := range claim.Messages() {
		if err := h.handler(msg.Value); err != nil {
			h.logger.Error("Failed to handle message").With("topic", msg.Topic).Err(err).Send()
		}
		session.MarkMessage(msg, "")
	}
	return nil
}
