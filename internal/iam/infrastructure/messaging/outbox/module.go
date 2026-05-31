package outbox

import (
	"github.com/tuannm99/podzone/internal/iam"
	"github.com/tuannm99/podzone/pkg/messaging"
	messagingkafka "github.com/tuannm99/podzone/pkg/messaging/kafka"
	"github.com/tuannm99/podzone/pkg/pdkafka"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/pdworker"
	"go.uber.org/fx"
)

var Module = fx.Options(
	iam.Module,
	fx.Provide(
		fx.Annotate(
			func(producer pdkafka.Producer) messaging.Publisher {
				return messagingkafka.NewPublisher(producer)
			},
			fx.ParamTags(`name:"kafka-iam-producer"`),
		),
		func(store messaging.OutboxStore, publisher messaging.Publisher) *messagingkafka.Relay {
			return messagingkafka.NewRelay(store, publisher, 100)
		},
		NewOutboxWorker,
	),
	fx.Invoke(func(lc fx.Lifecycle, logger pdlog.Logger, w *OutboxWorker) {
		pdworker.StartWorker(lc, logger, w)
	}),
)
