package pdkafka

import "github.com/IBM/sarama"

type Admin interface {
	DescribeCluster() (brokers []*sarama.Broker, controllerID int32, err error)
	Close() error
}
