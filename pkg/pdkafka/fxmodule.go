package pdkafka

import (
	"github.com/IBM/sarama"
	"go.uber.org/fx"
)

func ModuleFor(name string) fx.Option {
	if name == "" {
		name = "default"
	}

	nameParamTag := `name:"pdkafka-` + name + `"`
	configResultTag := `name:"kafka-` + name + `-config"`
	saramaConfigResultTag := `name:"kafka-` + name + `-sarama-config"`
	producerResultTag := `name:"kafka-` + name + `-producer"`
	adminResultTag := `name:"kafka-` + name + `-admin"`
	factoryResultTag := `name:"kafka-` + name + `-consumer-group-factory"`

	return fx.Options(
		fx.Supply(fx.Annotate(name, fx.ResultTags(nameParamTag))),
		fx.Provide(
			fx.Annotate(
				GetConfigFromKoanf,
				fx.ParamTags(nameParamTag, ``),
				fx.ResultTags(configResultTag),
			),
			fx.Annotate(
				NewSaramaConfig,
				fx.ParamTags(configResultTag),
				fx.ResultTags(saramaConfigResultTag),
			),
			fx.Annotate(
				NewSyncProducerFromConfig,
				fx.ParamTags(configResultTag, saramaConfigResultTag),
				fx.ResultTags(producerResultTag),
			),
			fx.Annotate(
				NewClusterAdminFromConfig,
				fx.ParamTags(configResultTag, saramaConfigResultTag),
				fx.ResultTags(adminResultTag),
			),
			fx.Annotate(
				NewConsumerGroupFactory,
				fx.ParamTags(configResultTag, saramaConfigResultTag),
				fx.ResultTags(factoryResultTag),
			),
			fx.Annotate(
				func(p sarama.SyncProducer) Producer { return p },
				fx.ParamTags(producerResultTag),
				fx.ResultTags(producerResultTag),
			),
			fx.Annotate(
				func(a sarama.ClusterAdmin) Admin { return a },
				fx.ParamTags(adminResultTag),
				fx.ResultTags(adminResultTag),
			),
		),
		fx.Invoke(
			fx.Annotate(
				registerLifecycle,
				fx.ParamTags(``, producerResultTag, adminResultTag, ``, configResultTag),
			),
		),
	)
}
