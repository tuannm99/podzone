package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/IBM/sarama"

	"github.com/tuannm99/podzone/pkg/messaging"
	"github.com/tuannm99/podzone/pkg/pdkafka"
)

type Publisher struct {
	producer pdkafka.Producer
}

var _ messaging.Publisher = (*Publisher)(nil)

func NewPublisher(producer pdkafka.Producer) *Publisher {
	return &Publisher{producer: producer}
}

func (p *Publisher) Publish(ctx context.Context, topic string, key string, msg messaging.Envelope) error {
	if err := msg.Validate(); err != nil {
		return err
	}
	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal envelope: %w", err)
	}

	_, _, err = p.producer.SendMessage(&sarama.ProducerMessage{
		Topic:   topic,
		Key:     sarama.StringEncoder(key),
		Value:   sarama.ByteEncoder(payload),
		Headers: pdkafka.ToRecordHeaders(msg.Headers),
	})
	if err != nil {
		return fmt.Errorf("publish kafka message: %w", err)
	}
	return nil
}

func (p *Publisher) PublishBatch(ctx context.Context, topic string, msgs []messaging.PublishRequest) error {
	producerMessages := make([]*sarama.ProducerMessage, 0, len(msgs))
	for _, item := range msgs {
		if err := item.Msg.Validate(); err != nil {
			return err
		}
		payload, err := json.Marshal(item.Msg)
		if err != nil {
			return fmt.Errorf("marshal envelope: %w", err)
		}
		targetTopic := item.Topic
		if targetTopic == "" {
			targetTopic = topic
		}
		producerMessages = append(producerMessages, &sarama.ProducerMessage{
			Topic:   targetTopic,
			Key:     sarama.StringEncoder(item.Key),
			Value:   sarama.ByteEncoder(payload),
			Headers: pdkafka.ToRecordHeaders(item.Msg.Headers),
		})
	}
	if err := p.producer.SendMessages(producerMessages); err != nil {
		return fmt.Errorf("publish kafka messages: %w", err)
	}
	return nil
}
