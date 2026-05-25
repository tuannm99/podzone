package messaging

import (
	"time"

	"github.com/knadh/koanf/v2"
)

type ObservabilityConfig struct {
	Enabled        bool `koanf:"enabled"           mapstructure:"enabled"`
	LogHandled     bool `koanf:"log_handled"       mapstructure:"log_handled"`
	LogRetries     bool `koanf:"log_retries"       mapstructure:"log_retries"`
	LogDeadLetters bool `koanf:"log_dead_letters"  mapstructure:"log_dead_letters"`
	LogDrops       bool `koanf:"log_drops"         mapstructure:"log_drops"`
}

type IdempotencyConfig struct {
	Enabled      bool   `koanf:"enabled"       mapstructure:"enabled"`
	ConsumerName string `koanf:"consumer_name" mapstructure:"consumer_name"`
	TableName    string `koanf:"table_name"    mapstructure:"table_name"`
}

type ConsumerRuntimeConfig struct {
	Enabled       bool                `koanf:"enabled"         mapstructure:"enabled"`
	MaxAttempts   int                 `koanf:"max_attempts"    mapstructure:"max_attempts"`
	BaseDelay     time.Duration       `koanf:"base_delay"      mapstructure:"base_delay"`
	Observability ObservabilityConfig `koanf:"observability"   mapstructure:"observability"`
	Idempotency   IdempotencyConfig   `koanf:"idempotency"     mapstructure:"idempotency"`
}

type TopicBootstrapConfig struct {
	Enabled            bool     `koanf:"enabled"             mapstructure:"enabled"`
	MainTopics         []string `koanf:"main_topics"         mapstructure:"main_topics"`
	RetryAttempts      []int    `koanf:"retry_attempts"      mapstructure:"retry_attempts"`
	CreateDeadLetter   bool     `koanf:"create_dead_letter"  mapstructure:"create_dead_letter"`
	DefaultPartitions  int32    `koanf:"default_partitions"  mapstructure:"default_partitions"`
	DefaultReplication int16    `koanf:"default_replication" mapstructure:"default_replication"`
}

func DefaultConsumerRuntimeConfig(consumerName string) ConsumerRuntimeConfig {
	return ConsumerRuntimeConfig{
		Enabled:     true,
		MaxAttempts: 5,
		BaseDelay:   time.Second,
		Observability: ObservabilityConfig{
			Enabled:        false,
			LogRetries:     true,
			LogDeadLetters: true,
		},
		Idempotency: IdempotencyConfig{
			Enabled:      false,
			ConsumerName: consumerName,
			TableName:    "message_inbox",
		},
	}
}

func LoadConsumerRuntimeConfig(k *koanf.Koanf, path string, defaults ConsumerRuntimeConfig) ConsumerRuntimeConfig {
	cfg := defaults
	if k != nil && path != "" {
		_ = k.Unmarshal(path, &cfg)
	}
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = defaults.MaxAttempts
	}
	if cfg.BaseDelay <= 0 {
		cfg.BaseDelay = defaults.BaseDelay
	}
	if cfg.Idempotency.ConsumerName == "" {
		cfg.Idempotency.ConsumerName = defaults.Idempotency.ConsumerName
	}
	if cfg.Idempotency.TableName == "" {
		cfg.Idempotency.TableName = defaults.Idempotency.TableName
	}
	return cfg
}

func (c TopicBootstrapConfig) Normalize() TopicBootstrapConfig {
	cfg := c
	if cfg.DefaultPartitions <= 0 {
		cfg.DefaultPartitions = 3
	}
	if cfg.DefaultReplication <= 0 {
		cfg.DefaultReplication = 1
	}
	return cfg
}
