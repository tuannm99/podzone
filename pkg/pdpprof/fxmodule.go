package pdpprof

import (
	"context"
	"net/http"

	// register pprof handlers on DefaultServeMux
	_ "net/http/pprof"

	"github.com/knadh/koanf/v2"
	"github.com/tuannm99/podzone/pkg/pdlog"
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
	srv := &http.Server{
		Addr:    cfg.Addr,
		Handler: nil,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if !cfg.Enable {
				log.Info("pprof disabled")
				return nil
			}

			log.Info("Starting pprof server", "addr", cfg.Addr)

			go func() {
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Error("pprof server error", "err", err)
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			if !cfg.Enable {
				return nil
			}
			log.Info("Shutting down pprof server", "addr", cfg.Addr)
			if err := srv.Shutdown(ctx); err != nil {
				log.Error("pprof shutdown error", "err", err)
				return err
			}
			log.Info("pprof server stopped")
			return nil
		},
	})
}
