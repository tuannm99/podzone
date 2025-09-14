package kafkafx

import (
	"context"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/tuannm99/podzone/pkg/pdlog"
)

// ExactlyOnceClient implements exactly-once delivery pattern
type ExactlyOnceClient struct {
	client   *KafkaClient
	logger   pdlog.Logger
	config   *ExactlyOnceConfig
	producer sarama.SyncProducer
	consumer sarama.Consumer
}

func NewExactlyOnceClient(
	client *KafkaClient,
	config *ExactlyOnceConfig,
	logger pdlog.Logger,
) (*ExactlyOnceClient, error) {
	saramaConfig := config.ToSaramaConfig()

	producer, err := sarama.NewSyncProducer(config.Brokers, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	consumer, err := sarama.NewConsumer(config.Brokers, saramaConfig)
	if err != nil {
		producer.Close()
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	return &ExactlyOnceClient{
		client:   client,
		logger:   logger,
		config:   config,
		producer: producer,
		consumer: consumer,
	}, nil
}

func (c *ExactlyOnceClient) Send(ctx context.Context, topic string, message []byte) error {
	if err := c.producer.BeginTxn(); err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(message),
	}

	if _, _, err := c.producer.SendMessage(msg); err != nil {
		c.producer.AbortTxn()
		return fmt.Errorf("failed to send message: %w", err)
	}

	if err := c.producer.CommitTxn(); err != nil {
		c.producer.AbortTxn()
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (c *ExactlyOnceClient) Receive(ctx context.Context, topic string, handler func([]byte) error) error {
	partitionConsumer, err := c.consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
	if err != nil {
		return fmt.Errorf("failed to consume partition: %w", err)
	}

	go func() {
		defer partitionConsumer.Close()

		for {
			select {
			case msg := <-partitionConsumer.Messages():
				if err := handler(msg.Value); err != nil {
					c.logger.Error("Failed to handle message").With("topic", topic).Err(err).Send()
				}
			case err := <-partitionConsumer.Errors():
				c.logger.Error("Failed to consume message").With("topic", topic).Err(err).Send()
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

func (c *ExactlyOnceClient) Close() error {
	if err := c.producer.Close(); err != nil {
		c.logger.Error("Failed to close producer").Err(err).Send()
	}
	if err := c.consumer.Close(); err != nil {
		c.logger.Error("Failed to close consumer").Err(err).Send()
	}
	return nil
}
