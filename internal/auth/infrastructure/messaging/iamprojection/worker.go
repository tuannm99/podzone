package iamprojection

import (
	"context"
	"fmt"

	controller "github.com/tuannm99/podzone/internal/auth/controller/eventhandler/iamprojection"
	"github.com/tuannm99/podzone/pkg/messaging"
	messagingkafka "github.com/tuannm99/podzone/pkg/messaging/kafka"
	"github.com/tuannm99/podzone/pkg/pdkafka"
	"github.com/tuannm99/podzone/pkg/pdlog"
)

type Worker struct {
	log      pdlog.Logger
	runner   pdkafka.ConsumerGroupRunner
	consumer *messagingkafka.Consumer
	enabled  bool
}

func NewWorker(
	log pdlog.Logger,
	runner pdkafka.ConsumerGroupRunner,
	publisher messaging.Publisher,
	inbox messaging.InboxStore,
	observer messaging.Observer,
	cfg messaging.ConsumerRuntimeConfig,
	handler *controller.Handler,
) *Worker {
	middlewares := make([]messaging.Middleware, 0, 1)
	if cfg.Idempotency.Enabled {
		middlewares = append(middlewares, messaging.IdempotentConsumer(inbox, cfg.Idempotency.ConsumerName, nil))
	}
	return &Worker{
		log:     log,
		runner:  runner,
		enabled: cfg.Enabled,
		consumer: messagingkafka.NewConsumerWithOptions(
			runner,
			messaging.TopicsWithRetry(messaging.EventTopic("iam"), cfg.MaxAttempts),
			handler,
			messagingkafka.ConsumerOptions{
				Publisher:        publisher,
				RetryPolicy:      messaging.RetryPolicy{MaxAttempts: cfg.MaxAttempts, BaseDelay: cfg.BaseDelay},
				DeadLetterPolicy: messaging.DeadLetterPolicy{Strategy: messaging.DefaultTopicStrategy()},
				Classifier:       messaging.DefaultErrorClassifier(),
				Middlewares:      middlewares,
				Observer:         observer,
				ConsumerName:     cfg.Idempotency.ConsumerName,
			},
		),
	}
}

func (w *Worker) Run(ctx context.Context) {
	if !w.enabled {
		w.log.Info("Auth IAM projection worker disabled")
		return
	}

	defer func() {
		if err := w.runner.Close(); err != nil {
			w.log.Error("Close auth IAM projection consumer failed", "error", err)
		}
	}()

	for ctx.Err() == nil {
		if err := w.consumer.Run(ctx); err != nil && ctx.Err() == nil {
			w.log.Error("Auth IAM projection consumer failed", "error", err)
		}
	}
}

func NewConsumerGroupRunner(
	factory pdkafka.ConsumerGroupFactory,
	cfg *pdkafka.Config,
) (pdkafka.ConsumerGroupRunner, error) {
	groupID := cfg.ConsumerGroupPrefix + ".iam-projection"
	group, err := factory.New(groupID)
	if err != nil {
		return nil, fmt.Errorf("create auth IAM projection consumer group: %w", err)
	}
	return pdkafka.NewConsumerGroupRunner(group), nil
}
