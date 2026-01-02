package kvstores

import (
	"fmt"

	"github.com/knadh/koanf/v2"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/toolkit/kvstores"
	"go.uber.org/fx"
)

// ConsulKVConfig holds Consul KV settings from koanf.
type ConsulKVConfig struct {
	Address string `mapstructure:"address"`
	Token   string `mapstructure:"token"`
	TLS     struct {
		InsecureSkipVerify bool `mapstructure:"insecure_skip_verify"`
	} `mapstructure:"tls"`
}

func NewConsulKVConfigFromKoanf(k *koanf.Koanf) (ConsulKVConfig, error) {
	var cfg ConsulKVConfig
	cfg.TLS.InsecureSkipVerify = true

	if k == nil {
		return cfg, fmt.Errorf("koanf is nil")
	}
	// Unmarshal from "consul_kv" prefix
	if err := k.Unmarshal("consul_kv", &cfg); err != nil {
		return cfg, fmt.Errorf("unmarshal consul config failed: %w", err)
	}
	if cfg.Address == "" {
		return cfg, fmt.Errorf("missing config: consul_kv.address")
	}
	return cfg, nil
}

func NewConsulKVStoreFromConfig(logger pdlog.Logger, cfg ConsulKVConfig) (*kvstores.ConsulKVStore, error) {
	return kvstores.NewConsulKVStore(logger, cfg.Address, cfg.Token, cfg.TLS.InsecureSkipVerify)
}

var Module = fx.Options(
	fx.Provide(
		NewConsulKVConfigFromKoanf,
		fx.Annotate(NewConsulKVStoreFromConfig, fx.As(new(kvstores.KVStore))),
	),
)
