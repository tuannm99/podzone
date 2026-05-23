package pdkafka

import (
	"context"

	"github.com/IBM/sarama"
)

type ConsumerGroupRunner interface {
	Run(ctx context.Context, topics []string, handler sarama.ConsumerGroupHandler) error
	Close() error
}

type consumerGroupRunner struct {
	group sarama.ConsumerGroup
}

func NewConsumerGroupRunner(group sarama.ConsumerGroup) ConsumerGroupRunner {
	return &consumerGroupRunner{group: group}
}

func (r *consumerGroupRunner) Run(ctx context.Context, topics []string, handler sarama.ConsumerGroupHandler) error {
	return r.group.Consume(ctx, topics, handler)
}

func (r *consumerGroupRunner) Close() error {
	return r.group.Close()
}
