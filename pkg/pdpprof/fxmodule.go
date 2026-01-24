package pdpprof

import (
	"context"

	// register pprof handlers on DefaultServeMux
	_ "net/http/pprof"

	"github.com/knadh/koanf/v2"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/pdserver"
	"go.uber.org/fx"
)

type Config struct {
	Enable bool   `koanf:"enable" mapstructure:"enable" yaml:"enable"`
	Addr   string `koanf:"addr"   mapstructure:"addr"   yaml:"addr"`
}

func NewConfigFromKoanf(k *koanf.Koanf) *Config {
	cfg := &Config{
		Enable: false,
		Addr:   ":6060",
	}

	// Unmarshal from "pprof" section if present
	if k != nil && k.Exists("pprof") {
		_ = k.Unmarshal("pprof", cfg)
	}

	if cfg.Addr == "" {
		cfg.Addr = ":6060"
	}

	return cfg
}

var Module = fx.Options(
	fx.Provide(
		NewConfigFromKoanf,
	),
	fx.Invoke(
		registerLifecycle,
	),
)

func registerLifecycle(lc fx.Lifecycle, log pdlog.Logger, cfg *Config) {
	if !cfg.Enable {
		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				log.Info("pprof disabled")
				return nil
			},
		})
		return
	}

	pdserver.RegisterHTTPServer(lc, log, cfg.Addr, nil, pdserver.WithComponent("pprof"))
}
