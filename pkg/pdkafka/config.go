package pdkafka

import (
	"fmt"

	"github.com/knadh/koanf/v2"
	"github.com/tuannm99/podzone/pkg/messaging"
)

type (
	RequiredAcks string
	Compression  string
)

const (
	RequiredAcksAll   RequiredAcks = "all"
	RequiredAcksLocal RequiredAcks = "local"
	RequiredAcksNone  RequiredAcks = "none"
	CompressionNone   Compression  = "none"
	CompressionGZIP   Compression  = "gzip"
	CompressionSnappy Compression  = "snappy"
	CompressionLZ4    Compression  = "lz4"
	CompressionZSTD   Compression  = "zstd"
)

type SASLConfig struct {
	Enabled   bool   `koanf:"enabled"   mapstructure:"enabled"`
	Mechanism string `koanf:"mechanism" mapstructure:"mechanism"`
	Username  string `koanf:"username"  mapstructure:"username"`
	Password  string `koanf:"password"  mapstructure:"password"`
}

type TLSConfig struct {
	Enabled bool `koanf:"enabled" mapstructure:"enabled"`
}

type Config struct {
	Brokers             []string                       `koanf:"brokers"               mapstructure:"brokers"`
	ClientID            string                         `koanf:"client_id"             mapstructure:"client_id"`
	Version             string                         `koanf:"version"               mapstructure:"version"`
	RequiredAcks        RequiredAcks                   `koanf:"required_acks"         mapstructure:"required_acks"`
	Compression         Compression                    `koanf:"compression"           mapstructure:"compression"`
	AutoCreateTopics    bool                           `koanf:"auto_create_topics"    mapstructure:"auto_create_topics"`
	ConsumerGroupPrefix string                         `koanf:"consumer_group_prefix" mapstructure:"consumer_group_prefix"`
	Topics              messaging.TopicBootstrapConfig `koanf:"topics"                mapstructure:"topics"`
	ProducerIdempotent  *bool                          `koanf:"producer_idempotent"   mapstructure:"producer_idempotent"`
	ProducerRetryMax    int                            `koanf:"producer_retry_max"    mapstructure:"producer_retry_max"`
	NetMaxOpenRequests  int                            `koanf:"net_max_open_requests" mapstructure:"net_max_open_requests"`
	RebalanceStrategy   string                         `koanf:"rebalance_strategy"    mapstructure:"rebalance_strategy"`
	SASL                SASLConfig                     `koanf:"sasl"                  mapstructure:"sasl"`
	TLS                 TLSConfig                      `koanf:"tls"                   mapstructure:"tls"`
}

func GetConfigFromKoanf(name string, k *koanf.Koanf) (*Config, error) {
	base := "kafka." + name
	messagingTopicsPath := "messaging.kafka." + name + ".topics"

	var cfg Config
	if err := k.Unmarshal(base, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal %q: %w", base, err)
	}
	if len(cfg.Brokers) == 0 {
		return nil, fmt.Errorf("kafka %q brokers is required", name)
	}
	if cfg.ClientID == "" {
		cfg.ClientID = "podzone-" + name
	}
	if cfg.ConsumerGroupPrefix == "" {
		cfg.ConsumerGroupPrefix = "podzone." + name
	}
	if cfg.Version == "" {
		cfg.Version = "3.7.0"
	}
	if cfg.RequiredAcks == "" {
		cfg.RequiredAcks = RequiredAcksAll
	}
	if cfg.Compression == "" {
		cfg.Compression = CompressionZSTD
	}
	if cfg.ProducerIdempotent == nil {
		v := true
		cfg.ProducerIdempotent = &v
	}
	if cfg.ProducerRetryMax <= 0 {
		cfg.ProducerRetryMax = 5
	}
	if cfg.NetMaxOpenRequests <= 0 {
		cfg.NetMaxOpenRequests = 1
	}
	if cfg.RebalanceStrategy == "" {
		cfg.RebalanceStrategy = "range"
	}
	if k != nil {
		topicsCfg := cfg.Topics
		_ = k.Unmarshal(messagingTopicsPath, &topicsCfg)
		cfg.Topics = topicsCfg
	}
	cfg.Topics = cfg.Topics.Normalize()
	return &cfg, nil
}
