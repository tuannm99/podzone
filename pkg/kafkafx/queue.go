package kafkafx

import (
	"context"
	"fmt"

	"github.com/IBM/sarama"
	"go.uber.org/zap"
)

// QueueClient implements FIFO queue pattern
type QueueClient struct {
	client   *KafkaClient
	logger   *zap.Logger
	config   *QueueConfig
	producer sarama.SyncProducer
	consumer sarama.Consumer
}

func NewQueueClient(client *KafkaClient, config *QueueConfig, logger *zap.Logger) (*QueueClient, error) {
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

	return &QueueClient{
		client:   client,
		logger:   logger,
		config:   config,
		producer: producer,
		consumer: consumer,
	}, nil
}

func (c *QueueClient) Enqueue(ctx context.Context, queue string, message []byte) error {
	_, _, err := c.producer.SendMessage(&sarama.ProducerMessage{
		Topic: queue,
		Value: sarama.ByteEncoder(message),
	})
	return err
}

func (c *QueueClient) Dequeue(ctx context.Context, queue string, handler func([]byte) error) error {
	partitionConsumer, err := c.consumer.ConsumePartition(queue, 0, sarama.OffsetOldest)
	if err != nil {
		return fmt.Errorf("failed to consume partition: %w", err)
	}

	go func() {
		defer partitionConsumer.Close()

		for {
			select {
			case msg := <-partitionConsumer.Messages():
				if err := handler(msg.Value); err != nil {
					c.logger.Error("Failed to handle message",
						zap.String("queue", queue),
						zap.Error(err))
				}
			case err := <-partitionConsumer.Errors():
				c.logger.Error("Failed to consume message",
					zap.String("queue", queue),
					zap.Error(err))
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

func (c *QueueClient) Close() error {
	if err := c.producer.Close(); err != nil {
		c.logger.Error("Failed to close producer", zap.Error(err))
	}
	if err := c.consumer.Close(); err != nil {
		c.logger.Error("Failed to close consumer", zap.Error(err))
	}
	return nil
}
