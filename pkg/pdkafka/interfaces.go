package pdkafka

import (
	"context"

	"github.com/IBM/sarama"
)

type Producer interface {
	SendMessage(msg *sarama.ProducerMessage) (partition int32, offset int64, err error)
	SendMessages(msgs []*sarama.ProducerMessage) error
	Close() error
}

type Admin interface {
	DescribeCluster() (brokers []*sarama.Broker, controllerID int32, err error)
	CreateTopic(topic string, detail *sarama.TopicDetail, validateOnly bool) error
	ListTopics() (map[string]sarama.TopicDetail, error)
	Close() error
}

type ConsumerGroupFactory interface {
	New(groupID string) (sarama.ConsumerGroup, error)
}

type ConsumerGroupRunner interface {
	Run(ctx context.Context, topics []string, handler sarama.ConsumerGroupHandler) error
	Close() error
}
