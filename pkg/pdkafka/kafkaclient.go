package pdkafka

// import (
// 	"context"
// 	"fmt"
// 	"sync"
//
// 	"github.com/IBM/sarama"
// 	pdlog "github.com/tuannm99/podzone/pkg/pdlog"
// )
//
// // KafkaClient represents a Kafka client
// type KafkaClient struct {
// 	producer sarama.SyncProducer
// 	consumer sarama.Consumer
// 	config   *Config
// 	logger   pdlog.Logger
// 	mu       sync.RWMutex
// }
//
// // Config holds Kafka connection configuration
// type Config struct {
// 	Brokers []string
// 	GroupID string
// }
//
// // NewKafkaClient creates a new Kafka client
// func NewKafkaClient(config *Config, logger pdlog.Logger) (*KafkaClient, error) {
// 	// Producer config
// 	producerConfig := sarama.NewConfig()
// 	producerConfig.Producer.RequiredAcks = sarama.WaitForAll
// 	producerConfig.Producer.Retry.Max = 5
// 	producerConfig.Producer.Return.Successes = true
//
// 	// Create producer
// 	producer, err := sarama.NewSyncProducer(config.Brokers, producerConfig)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create producer: %w", err)
// 	}
//
// 	// Consumer config
// 	consumerConfig := sarama.NewConfig()
// 	consumerConfig.Consumer.Return.Errors = true
//
// 	// Create consumer
// 	consumer, err := sarama.NewConsumer(config.Brokers, consumerConfig)
// 	if err != nil {
// 		producer.Close()
// 		return nil, fmt.Errorf("failed to create consumer: %w", err)
// 	}
//
// 	return &KafkaClient{
// 		producer: producer,
// 		consumer: consumer,
// 		config:   config,
// 		logger:   logger,
// 	}, nil
// }
//
// // PublishMessage publishes a message to a Kafka topic
// func (c *KafkaClient) PublishMessage(ctx context.Context, topic string, key, value []byte) error {
// 	c.mu.RLock()
// 	defer c.mu.RUnlock()
//
// 	msg := &sarama.ProducerMessage{
// 		Topic: topic,
// 		Key:   sarama.ByteEncoder(key),
// 		Value: sarama.ByteEncoder(value),
// 	}
//
// 	_, _, err := c.producer.SendMessage(msg)
// 	if err != nil {
// 		return fmt.Errorf("failed to send message: %w", err)
// 	}
//
// 	return nil
// }
//
// // Subscribe subscribes to a Kafka topic
// func (c *KafkaClient) Subscribe(ctx context.Context, topic string, handler func(key, value []byte) error) error {
// 	c.mu.RLock()
// 	defer c.mu.RUnlock()
//
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
// 				if err := handler(msg.Key, msg.Value); err != nil {
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
// // Close closes the Kafka client
// func (c *KafkaClient) Close() error {
// 	c.mu.Lock()
// 	defer c.mu.Unlock()
//
// 	if err := c.producer.Close(); err != nil {
// 		return fmt.Errorf("failed to close producer: %w", err)
// 	}
//
// 	if err := c.consumer.Close(); err != nil {
// 		return fmt.Errorf("failed to close consumer: %w", err)
// 	}
//
// 	return nil
// }
