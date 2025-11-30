package pdpprof

import (
	"context"
	"net/http"

	// register pprof handlers on DefaultServeMux
	_ "net/http/pprof"

	"github.com/spf13/viper"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"go.uber.org/fx"
)

type Config struct {
	Enable bool   `mapstructure:"enable" yaml:"enable"`
	Addr   string `mapstructure:"addr"   yaml:"addr"`
}

func NewConfigFromViper(v *viper.Viper) *Config {
	cfg := &Config{
		Enable: v.GetBool("pprof_enable"),
		Addr:   v.GetString("pprof_addr"),
	}

	if sub := v.Sub("pprof"); sub != nil {
		_ = sub.Unmarshal(cfg)
	}

	if cfg.Addr == "" {
		cfg.Addr = ":6060"
	}

	return cfg
}

var Module = fx.Options(
	fx.Provide(
		NewConfigFromViper,
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
