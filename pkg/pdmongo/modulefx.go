package pdmongo

import (
	"context"
	"time"

	"github.com/tuannm99/podzone/pkg/pdlog"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/fx"
)

func ModuleFor(name string) fx.Option {
	if name == "" {
		name = "default"
	}

	nameParamTag := `name:"pdmongo-` + name + `"`
	configResultTag := `name:"mongo-` + name + `-config"`
	clientResultTag := `name:"mongo-` + name + `"`

	return fx.Options(
		fx.Supply(
			fx.Annotate(name, fx.ResultTags(nameParamTag)),
		),
		fx.Provide(
			fx.Annotate(
				GetConfigFromViper,
				fx.ParamTags(nameParamTag, ``),
				fx.ResultTags(configResultTag),
			),
			fx.Annotate(
				NewMongoDbFromConfig,
				fx.ParamTags(configResultTag),
				fx.ResultTags(clientResultTag),
			),
		),
		fx.Invoke(
			fx.Annotate(
				registerLifecycle,
				fx.ParamTags(``, clientResultTag, ``, configResultTag),
			),
		),
	)
}

func registerLifecycle(lc fx.Lifecycle, client *mongo.Client, log pdlog.Logger, cfg *Config) {
	const (
		maxAttempts    = 5
		initialBackoff = 200 * time.Millisecond
		maxBackoff     = 2 * time.Second
	)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			backoff := initialBackoff
			var lastErr error
			for attempt := 1; attempt <= maxAttempts; attempt++ {
				log.Info("Pinging Mongo...", "uri", cfg.URI, "attempt", attempt, "max_attempts", maxAttempts)
				pingCtx, cancel := context.WithTimeout(ctx, cfg.PingTimeout)
				lastErr = client.Ping(pingCtx, nil)
				cancel()
				if lastErr == nil {
					log.Info("Mongo ping OK", "uri", cfg.URI)
					return nil
				}
				if ctx.Err() != nil {
					break
				}
				t := time.NewTimer(backoff)
				select {
				case <-ctx.Done():
					t.Stop()
					return ctx.Err()
				case <-t.C:
				}
				if backoff *= 2; backoff > maxBackoff {
					backoff = maxBackoff
				}
			}
			log.Error("Mongo ping failed", "error", lastErr, "uri", cfg.URI)
			return lastErr
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Closing Mongo connection", "uri", cfg.URI)
			if err := client.Disconnect(ctx); err != nil {
				log.Error("Close Mongo failed", "error", err)
				return err
			}
			log.Info("Mongo connection closed")
			return nil
		},
	})
}
