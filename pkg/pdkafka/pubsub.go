package pdkafka

// import (
// 	"context"
// 	"fmt"
//
// 	"github.com/IBM/sarama"
// 	pdlog "github.com/tuannm99/podzone/pkg/pdlog"
// )
//
// // PubSubClient implements publish-subscribe pattern
// type PubSubClient struct {
// 	client   *KafkaClient
// 	logger   pdlog.Logger
// 	config   *PubSubConfig
// 	producer sarama.SyncProducer
// 	consumer sarama.Consumer
// }
//
// func NewPubSubClient(client *KafkaClient, config *PubSubConfig, logger pdlog.Logger) (*PubSubClient, error) {
// 	saramaConfig := config.ToSaramaConfig()
//
// 	producer, err := sarama.NewSyncProducer(config.Brokers, saramaConfig)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create producer: %w", err)
// 	}
//
// 	consumer, err := sarama.NewConsumer(config.Brokers, saramaConfig)
// 	if err != nil {
// 		producer.Close()
// 		return nil, fmt.Errorf("failed to create consumer: %w", err)
// 	}
//
// 	return &PubSubClient{
// 		client:   client,
// 		logger:   logger,
// 		config:   config,
// 		producer: producer,
// 		consumer: consumer,
// 	}, nil
// }
//
// func (c *PubSubClient) Publish(ctx context.Context, topic string, message []byte) error {
// 	_, _, err := c.producer.SendMessage(&sarama.ProducerMessage{
// 		Topic: topic,
// 		Value: sarama.ByteEncoder(message),
// 	})
// 	return err
// }
//
// func (c *PubSubClient) Subscribe(ctx context.Context, topic string, handler func([]byte) error) error {
// 	partitionConsumer, err := c.consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
// 	if err != nil {
// 		return fmt.Errorf("failed to consume partition: %w", err)
// 	}
//
// 	go func() {
// 		defer partitionConsumer.Close()
//
// 		for {
// 			select {
// 			case msg := <-partitionConsumer.Messages():
// 				if err := handler(msg.Value); err != nil {
// 					c.logger.Error("Failed to handle message").With("topic", topic).Err(err).Send()
// 				}
// 			case err := <-partitionConsumer.Errors():
// 				c.logger.Error("Failed to consume message").With("topic", topic).Err(err).Send()
// 			case <-ctx.Done():
// 				return
// 			}
// 		}
// 	}()
//
// 	return nil
// }
//
// func (c *PubSubClient) Close() error {
// 	if err := c.producer.Close(); err != nil {
// 		c.logger.Error("Failed to close producer").Err(err).Send()
// 	}
// 	if err := c.consumer.Close(); err != nil {
// 		c.logger.Error("Failed to close consumer").Err(err).Send()
// 	}
// 	return nil
// }
