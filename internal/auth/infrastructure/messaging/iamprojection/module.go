package iamprojection

import (
	"github.com/jmoiron/sqlx"
	"github.com/knadh/koanf/v2"
	"go.uber.org/fx"

	controller "github.com/tuannm99/podzone/internal/auth/controller/eventhandler/iamprojection"
	"github.com/tuannm99/podzone/internal/auth/domain/outputport"
	"github.com/tuannm99/podzone/internal/auth/infrastructure/repository"
	"github.com/tuannm99/podzone/pkg/messaging"
	messagingkafka "github.com/tuannm99/podzone/pkg/messaging/kafka"
	messagingsql "github.com/tuannm99/podzone/pkg/messaging/sqlstore"
	"github.com/tuannm99/podzone/pkg/pdkafka"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/pdworker"
)

const (
	runtimeConfigPath = "messaging.auth.consumers.iam_projection"
	consumerName      = "auth.iam-projection"
)

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(repository.NewIAMProjectionRepositoryImpl, fx.As(new(outputport.IAMProjectionRepository))),
		fx.Annotate(NewRuntimeConfig, fx.ResultTags(`name:"auth-iam-projection-runtime"`)),
		fx.Annotate(
			NewConsumerGroupRunner,
			fx.ParamTags(`name:"kafka-auth-consumer-group-factory"`, `name:"kafka-auth-config"`),
		),
		fx.Annotate(
			func(producer pdkafka.Producer) messaging.Publisher {
				return messagingkafka.NewPublisher(producer)
			},
			fx.ParamTags(`name:"kafka-auth-producer"`),
		),
		fx.Annotate(
			func(db *sqlx.DB, cfg messaging.ConsumerRuntimeConfig) (messaging.InboxStore, error) {
				return messagingsql.NewInboxStore(db, cfg.Idempotency.TableName)
			},
			fx.ParamTags(`name:"sql-auth"`, `name:"auth-iam-projection-runtime"`),
		),
		fx.Annotate(
			func(log pdlog.Logger, cfg messaging.ConsumerRuntimeConfig) messaging.Observer {
				return messaging.NewLoggingObserver(log, consumerName, cfg)
			},
			fx.ParamTags(``, `name:"auth-iam-projection-runtime"`),
		),
		controller.NewHandler,
		fx.Annotate(
			NewWorker,
			fx.ParamTags(``, ``, ``, ``, ``, `name:"auth-iam-projection-runtime"`, ``),
		),
	),
	fx.Invoke(func(lc fx.Lifecycle, logger pdlog.Logger, w *Worker) {
		pdworker.StartWorker(lc, logger, w)
	}),
)

func NewRuntimeConfig(k *koanf.Koanf) messaging.ConsumerRuntimeConfig {
	cfg := messaging.LoadConsumerRuntimeConfig(
		k,
		runtimeConfigPath,
		messaging.DefaultConsumerRuntimeConfig(consumerName),
	)
	cfg.Idempotency.ConsumerName = consumerName
	return cfg
}
