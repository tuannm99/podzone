package pdsql

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"go.uber.org/fx"
)

func ModuleFor(name string) fx.Option {
	if name == "" {
		name = "default"
	}
	nameTag := `name:"pdsql` + name + `"`
	resultTag := `name:"sql-` + name + `"`

	return fx.Options(
		fx.Supply(
			// provide the name string into the container
			fx.Annotate(name, fx.ResultTags(nameTag)),
		),
		fx.Provide(
			fx.Annotate(GetConfigFromViper, fx.ParamTags(nameTag)), // inject string[name="<name>"]
			fx.Annotate(NewDbFromConfig, fx.ResultTags(resultTag)), // provides *sqlx.DB[name="sql-<name>"]
		),
		fx.Invoke(
			fx.Annotate(
				registerLifecycle,
				// params: lc, *sqlx.DB(named), pdlog.Logger, *Config
				fx.ParamTags(``, resultTag, ``, ``),
			),
		),
	)
}

func registerLifecycle(lc fx.Lifecycle, db *sqlx.DB, log pdlog.Logger, cfg *Config) {
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
				log.Info("Pinging database...", "uri", cfg.URI, "attempt", attempt, "max_attempts", maxAttempts)
				if lastErr = db.PingContext(ctx); lastErr == nil {
					log.Info("Database ping OK", "uri", cfg.URI)
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
			log.Error("Database ping failed", "error", lastErr, "uri", cfg.URI)
			return lastErr
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Closing DB connection", "uri", cfg.URI)
			if db == nil {
				return nil
			}
			if err := db.Close(); err != nil {
				log.Error("Close DB failed", "error", err)
				return err
			}
			log.Info("DB connection closed")
			return nil
		},
	})
}
