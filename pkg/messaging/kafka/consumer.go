package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/IBM/sarama"

	"github.com/tuannm99/podzone/pkg/messaging"
	"github.com/tuannm99/podzone/pkg/pdkafka"
)

type Consumer struct {
	runner  pdkafka.ConsumerGroupRunner
	topics  []string
	handler messaging.Handler
}

var _ messaging.Consumer = (*Consumer)(nil)

func NewConsumer(runner pdkafka.ConsumerGroupRunner, topics []string, handler messaging.Handler) *Consumer {
	return &Consumer{
		runner:  runner,
		topics:  append([]string(nil), topics...),
		handler: handler,
	}
}

func (c *Consumer) Run(ctx context.Context) error {
	return c.runner.Run(ctx, c.topics, &consumerGroupHandler{handler: c.handler})
}

type consumerGroupHandler struct {
	handler messaging.Handler
}

func (h *consumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *consumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (h *consumerGroupHandler) ConsumeClaim(
	session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim,
) error {
	for {
		select {
		case <-session.Context().Done():
			return nil
		case msg, ok := <-claim.Messages():
			if !ok {
				return nil
			}
			var env messaging.Envelope
			if err := json.Unmarshal(msg.Value, &env); err != nil {
				return fmt.Errorf("unmarshal envelope: %w", err)
			}
			if err := h.handler.Handle(session.Context(), env); err != nil {
				return err
			}
			session.MarkMessage(msg, "")
		}
	}
}
