package pdkafka

import (
	"crypto/tls"
	"fmt"
	"strings"

	"github.com/IBM/sarama"
)

type ConsumerGroupFactory interface {
	New(groupID string) (sarama.ConsumerGroup, error)
}

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

	version, err := sarama.ParseKafkaVersion(cfg.Version)
	if err != nil {
		return nil, fmt.Errorf("parse kafka version %q: %w", cfg.Version, err)
	}

	scfg := sarama.NewConfig()
	scfg.Version = version
	scfg.ClientID = cfg.ClientID
	scfg.Metadata.Full = true
	scfg.Metadata.AllowAutoTopicCreation = cfg.AutoCreateTopics

	scfg.Producer.Return.Successes = true
	scfg.Producer.Return.Errors = true
	scfg.Producer.Idempotent = true
	scfg.Producer.Retry.Max = 5
	scfg.Producer.Partitioner = sarama.NewHashPartitioner
	scfg.Producer.RequiredAcks = mapRequiredAcks(cfg.RequiredAcks)
	scfg.Producer.Compression = mapCompression(cfg.Compression)
	scfg.Net.MaxOpenRequests = 1

	scfg.Consumer.Return.Errors = true
	scfg.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRange()
	scfg.Consumer.Offsets.Initial = sarama.OffsetNewest

	if cfg.SASL.Enabled {
		scfg.Net.SASL.Enable = true
		scfg.Net.SASL.User = cfg.SASL.Username
		scfg.Net.SASL.Password = cfg.SASL.Password
		switch strings.ToUpper(cfg.SASL.Mechanism) {
		case "SCRAM-SHA-256":
			scfg.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA256
		case "SCRAM-SHA-512":
			scfg.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA512
		case "PLAIN", "":
			scfg.Net.SASL.Mechanism = sarama.SASLTypePlaintext
		default:
			return nil, fmt.Errorf("unsupported kafka sasl mechanism %q", cfg.SASL.Mechanism)
		}
	}

	if cfg.TLS.Enabled {
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
