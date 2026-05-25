package pdkafka

import (
	"crypto/tls"
	"fmt"
	"strings"

	"github.com/IBM/sarama"
)

type consumerGroupFactory struct {
	brokers []string
	config  *sarama.Config
}

func (f consumerGroupFactory) New(groupID string) (sarama.ConsumerGroup, error) {
	if strings.TrimSpace(groupID) == "" {
		return nil, fmt.Errorf("consumer group id is required")
	}
	return sarama.NewConsumerGroup(f.brokers, groupID, f.config)
}

func NewSaramaConfig(cfg *Config) (*sarama.Config, error) {
	if cfg == nil {
		return nil, fmt.Errorf("nil kafka config")
	}
	normalized := *cfg
	if normalized.Version == "" {
		normalized.Version = "3.7.0"
	}
	if normalized.RequiredAcks == "" {
		normalized.RequiredAcks = RequiredAcksAll
	}
	if normalized.Compression == "" {
		normalized.Compression = CompressionZSTD
	}
	if normalized.ProducerIdempotent == nil {
		v := true
		normalized.ProducerIdempotent = &v
	}
	if normalized.ProducerRetryMax <= 0 {
		normalized.ProducerRetryMax = 5
	}
	if normalized.NetMaxOpenRequests <= 0 {
		normalized.NetMaxOpenRequests = 1
	}
	if normalized.RebalanceStrategy == "" {
		normalized.RebalanceStrategy = "range"
	}

	version, err := sarama.ParseKafkaVersion(normalized.Version)
	if err != nil {
		return nil, fmt.Errorf("parse kafka version %q: %w", cfg.Version, err)
	}

	scfg := sarama.NewConfig()
	scfg.Version = version
	scfg.ClientID = normalized.ClientID
	scfg.Metadata.Full = true
	scfg.Metadata.AllowAutoTopicCreation = normalized.AutoCreateTopics

	scfg.Producer.Return.Successes = true
	scfg.Producer.Return.Errors = true
	scfg.Producer.Idempotent = normalized.ProducerIdempotent != nil && *normalized.ProducerIdempotent
	scfg.Producer.Retry.Max = normalized.ProducerRetryMax
	scfg.Producer.Partitioner = sarama.NewHashPartitioner
	scfg.Producer.RequiredAcks = mapRequiredAcks(normalized.RequiredAcks)
	scfg.Producer.Compression = mapCompression(normalized.Compression)
	scfg.Net.MaxOpenRequests = normalized.NetMaxOpenRequests

	scfg.Consumer.Return.Errors = true
	strategy, err := mapRebalanceStrategy(normalized.RebalanceStrategy)
	if err != nil {
		return nil, err
	}
	scfg.Consumer.Group.Rebalance.Strategy = strategy
	scfg.Consumer.Offsets.Initial = sarama.OffsetNewest

	if normalized.SASL.Enabled {
		scfg.Net.SASL.Enable = true
		scfg.Net.SASL.User = normalized.SASL.Username
		scfg.Net.SASL.Password = normalized.SASL.Password
		switch strings.ToUpper(normalized.SASL.Mechanism) {
		case "SCRAM-SHA-256":
			scfg.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA256
		case "SCRAM-SHA-512":
			scfg.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA512
		case "PLAIN", "":
			scfg.Net.SASL.Mechanism = sarama.SASLTypePlaintext
		default:
			return nil, fmt.Errorf("unsupported kafka sasl mechanism %q", normalized.SASL.Mechanism)
		}
	}

	if normalized.TLS.Enabled {
		scfg.Net.TLS.Enable = true
		scfg.Net.TLS.Config = &tls.Config{MinVersion: tls.VersionTLS12}
	}

	return scfg, nil
}

func NewSyncProducerFromConfig(cfg *Config, scfg *sarama.Config) (sarama.SyncProducer, error) {
	if cfg == nil {
		return nil, fmt.Errorf("nil kafka config")
	}
	return sarama.NewSyncProducer(cfg.Brokers, scfg)
}

func NewClusterAdminFromConfig(cfg *Config, scfg *sarama.Config) (sarama.ClusterAdmin, error) {
	if cfg == nil {
		return nil, fmt.Errorf("nil kafka config")
	}
	return sarama.NewClusterAdmin(cfg.Brokers, scfg)
}

func NewConsumerGroupFactory(cfg *Config, scfg *sarama.Config) ConsumerGroupFactory {
	return consumerGroupFactory{
		brokers: append([]string(nil), cfg.Brokers...),
		config:  scfg,
	}
}

func mapRequiredAcks(v RequiredAcks) sarama.RequiredAcks {
	switch v {
	case RequiredAcksNone:
		return sarama.NoResponse
	case RequiredAcksLocal:
		return sarama.WaitForLocal
	default:
		return sarama.WaitForAll
	}
}

func mapCompression(v Compression) sarama.CompressionCodec {
	switch v {
	case CompressionGZIP:
		return sarama.CompressionGZIP
	case CompressionSnappy:
		return sarama.CompressionSnappy
	case CompressionLZ4:
		return sarama.CompressionLZ4
	case CompressionZSTD:
		return sarama.CompressionZSTD
	default:
		return sarama.CompressionNone
	}
}

func mapRebalanceStrategy(raw string) (sarama.BalanceStrategy, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "range":
		return sarama.NewBalanceStrategyRange(), nil
	case "round_robin", "roundrobin":
		return sarama.NewBalanceStrategyRoundRobin(), nil
	case "sticky":
		return sarama.NewBalanceStrategySticky(), nil
	default:
		return nil, fmt.Errorf("unsupported kafka rebalance strategy %q", raw)
	}
}
