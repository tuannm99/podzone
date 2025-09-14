package pdkafka

import (
	"time"

	"github.com/IBM/sarama"
)

// BaseConfig holds common Kafka configuration
type BaseConfig struct {
	Brokers []string
	Version sarama.KafkaVersion
}

// PubSubConfig holds configuration for pub/sub pattern
type PubSubConfig struct {
	BaseConfig
	// Consumer settings
	ConsumerGroupID string
	AutoCommit      bool
	// Producer settings
	RequiredAcks sarama.RequiredAcks
	RetryMax     int
}

// QueueConfig holds configuration for queue pattern
type QueueConfig struct {
	BaseConfig
	// Consumer settings
	ConsumerGroupID string
	AutoCommit      bool
	// Producer settings
	RequiredAcks sarama.RequiredAcks
	RetryMax     int
	// Queue settings
	Partitions int32
}

// ConsumerGroupConfig holds configuration for consumer group pattern
type ConsumerGroupConfig struct {
	BaseConfig
	GroupID            string
	Assignor           string // range, roundrobin, sticky
	Oldest             bool   // consume from oldest offset
	AutoCommit         bool
	AutoCommitInterval int
	RebalanceTimeout   int
	SessionTimeout     int
}

// ExactlyOnceConfig holds configuration for exactly-once pattern
type ExactlyOnceConfig struct {
	BaseConfig
	// Producer settings
	ProducerID   string
	RequiredAcks sarama.RequiredAcks
	RetryMax     int
	// Consumer settings
	ConsumerGroupID string
	IsolationLevel  sarama.IsolationLevel
	AutoCommit      bool
	// Transaction settings
	TransactionTimeout int
}

// DefaultConfigs returns default configurations for each pattern
func DefaultConfigs(brokers []string) map[string]interface{} {
	return map[string]interface{}{
		"pubsub": PubSubConfig{
			BaseConfig: BaseConfig{
				Brokers: brokers,
				Version: sarama.V2_8_0_0,
			},
			ConsumerGroupID: "pubsub-group",
			AutoCommit:      true,
			RequiredAcks:    sarama.WaitForAll,
			RetryMax:        5,
		},
		"queue": QueueConfig{
			BaseConfig: BaseConfig{
				Brokers: brokers,
				Version: sarama.V2_8_0_0,
			},
			ConsumerGroupID: "queue-group",
			AutoCommit:      true,
			RequiredAcks:    sarama.WaitForAll,
			RetryMax:        5,
			Partitions:      3,
		},
		"consumer_group": ConsumerGroupConfig{
			BaseConfig: BaseConfig{
				Brokers: brokers,
				Version: sarama.V2_8_0_0,
			},
			GroupID:            "consumer-group",
			Assignor:           "range",
			Oldest:             true,
			AutoCommit:         true,
			AutoCommitInterval: 1000,
			RebalanceTimeout:   60000,
			SessionTimeout:     30000,
		},
		"exactly_once": ExactlyOnceConfig{
			BaseConfig: BaseConfig{
				Brokers: brokers,
				Version: sarama.V2_8_0_0,
			},
			ProducerID:         "exactly-once-producer",
			RequiredAcks:       sarama.WaitForAll,
			RetryMax:           5,
			ConsumerGroupID:    "exactly-once-group",
			IsolationLevel:     sarama.ReadCommitted,
			AutoCommit:         false,
			TransactionTimeout: 60000,
		},
	}
}

// ToSaramaConfig converts pattern configs to Sarama configs
func (c *PubSubConfig) ToSaramaConfig() *sarama.Config {
	config := sarama.NewConfig()
	config.Version = c.Version
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRange()}
	config.Consumer.Offsets.AutoCommit.Enable = c.AutoCommit
	config.Producer.RequiredAcks = c.RequiredAcks
	config.Producer.Retry.Max = c.RetryMax
	return config
}

func (c *QueueConfig) ToSaramaConfig() *sarama.Config {
	config := sarama.NewConfig()
	config.Version = c.Version
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRange()}
	config.Consumer.Offsets.AutoCommit.Enable = c.AutoCommit
	config.Producer.RequiredAcks = c.RequiredAcks
	config.Producer.Retry.Max = c.RetryMax
	return config
}

func (c *ConsumerGroupConfig) ToSaramaConfig() *sarama.Config {
	config := sarama.NewConfig()
	config.Version = c.Version

	switch c.Assignor {
	case "sticky":
		config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategySticky()}
	case "roundrobin":
		config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{
			sarama.NewBalanceStrategyRoundRobin(),
		}
	case "range":
		config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRange()}
	}

	if c.Oldest {
		config.Consumer.Offsets.Initial = sarama.OffsetOldest
	}

	config.Consumer.Group.Session.Timeout = time.Duration(c.SessionTimeout) * time.Millisecond
	config.Consumer.Group.Rebalance.Timeout = time.Duration(c.RebalanceTimeout) * time.Millisecond
	config.Consumer.Offsets.AutoCommit.Enable = c.AutoCommit
	config.Consumer.Offsets.AutoCommit.Interval = time.Duration(c.AutoCommitInterval) * time.Millisecond

	return config
}

func (c *ExactlyOnceConfig) ToSaramaConfig() *sarama.Config {
	config := sarama.NewConfig()
	config.Version = c.Version
	config.Net.MaxOpenRequests = 1
	config.Producer.RequiredAcks = c.RequiredAcks
	config.Producer.Idempotent = true
	config.Producer.Transaction.ID = c.ProducerID
	config.Producer.Transaction.Timeout = time.Duration(c.TransactionTimeout) * time.Millisecond
	config.Producer.Retry.Max = c.RetryMax
	config.Consumer.IsolationLevel = c.IsolationLevel
	config.Consumer.Offsets.AutoCommit.Enable = c.AutoCommit
	return config
}
