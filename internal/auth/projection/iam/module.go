package iam

import (
	"go.uber.org/fx"

	"github.com/tuannm99/podzone/internal/auth/domain/outputport"
	"github.com/tuannm99/podzone/internal/auth/infrastructure/repository"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/pdworker"
)

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(repository.NewIAMProjectionRepositoryImpl, fx.As(new(outputport.IAMProjectionRepository))),
		fx.Annotate(
			NewConsumerGroupRunner,
			fx.ParamTags(`name:"kafka-auth-consumer-group-factory"`, `name:"kafka-auth-config"`),
		),
		NewHandler,
		NewWorker,
	),
	fx.Invoke(func(lc fx.Lifecycle, logger pdlog.Logger, w *Worker) {
		pdworker.StartWorker(lc, logger, w)
	}),
)
