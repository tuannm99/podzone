package postgresfx

import (
	"context"
	"fmt"

	"github.com/tuannm99/podzone/pkg/pdlog"
	"go.uber.org/fx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ModuleFor(name string, conStr string) fx.Option {
	pgName := fmt.Sprintf("%s-postgres-uri", name)
	resultName := fmt.Sprintf("gorm-%s", name)

	return fx.Options(
		fx.Provide(
			fx.Annotate(
				func() string { return conStr },
				fx.ResultTags(fmt.Sprintf(`name:"%s"`, pgName)),
			),
			fx.Annotate(
				NewGormDB,
				fx.ParamTags(``, ``, fmt.Sprintf(`name:"%s"`, pgName)),
				fx.ResultTags(fmt.Sprintf(`name:"%s"`, resultName)),
			),
		),
	)
}

func NewGormDB(lc fx.Lifecycle, logger pdlog.Logger, conStr string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(conStr), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("postgres connect failed: %w", err)
	}

	sqlDB, _ := db.DB()
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("Pinging Postgres...").With("dsn", conStr).Send()
			if err := sqlDB.PingContext(ctx); err != nil {
				return fmt.Errorf("postgres ping failed: %w", err)
			}
			logger.Info("Postgres is reachable").With("dsn", conStr).Send()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Closing Postgres connection")
			return sqlDB.Close()
		},
	})

	logger.Info("Connected to Postgres").With("dsn", conStr).Send()
	return db, nil
}
