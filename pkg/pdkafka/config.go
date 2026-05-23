package pdkafka

import (
	"fmt"

	"github.com/knadh/koanf/v2"
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
	Brokers             []string     `koanf:"brokers"               mapstructure:"brokers"`
	ClientID            string       `koanf:"client_id"             mapstructure:"client_id"`
	Version             string       `koanf:"version"               mapstructure:"version"`
	RequiredAcks        RequiredAcks `koanf:"required_acks"         mapstructure:"required_acks"`
	Compression         Compression  `koanf:"compression"           mapstructure:"compression"`
	AutoCreateTopics    bool         `koanf:"auto_create_topics"    mapstructure:"auto_create_topics"`
	ConsumerGroupPrefix string       `koanf:"consumer_group_prefix" mapstructure:"consumer_group_prefix"`
	SASL                SASLConfig   `koanf:"sasl"                  mapstructure:"sasl"`
	TLS                 TLSConfig    `koanf:"tls"                   mapstructure:"tls"`
}

func GetConfigFromKoanf(name string, k *koanf.Koanf) (*Config, error) {
	base := "kafka." + name

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
	return &cfg, nil
}
