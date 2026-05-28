package worker

import (
	"context"
	"fmt"

	"github.com/tuannm99/podzone/pkg/messaging"
	messagingkafka "github.com/tuannm99/podzone/pkg/messaging/kafka"
	"github.com/tuannm99/podzone/pkg/pdkafka"
	"github.com/tuannm99/podzone/pkg/pdlog"
)

type ConsumerWorker struct {
	log      pdlog.Logger
	runner   pdkafka.ConsumerGroupRunner
	consumer *messagingkafka.Consumer
	enabled  bool
}

func NewConsumerWorker(
	log pdlog.Logger,
	runner pdkafka.ConsumerGroupRunner,
	publisher messaging.Publisher,
	cfg messaging.ConsumerRuntimeConfig,
	observer messaging.Observer,
	handler messaging.Handler,
) *ConsumerWorker {
	return &ConsumerWorker{
		log:     log,
		runner:  runner,
		enabled: cfg.Enabled,
		consumer: messagingkafka.NewConsumerWithOptions(
			runner,
			messaging.TopicsWithRetry(messaging.EventTopic("onboarding"), cfg.MaxAttempts),
			handler,
			messagingkafka.ConsumerOptions{
				Publisher:        publisher,
				RetryPolicy:      messaging.RetryPolicy{MaxAttempts: cfg.MaxAttempts, BaseDelay: cfg.BaseDelay},
				DeadLetterPolicy: messaging.DeadLetterPolicy{Strategy: messaging.DefaultTopicStrategy()},
				Classifier:       messaging.DefaultErrorClassifier(),
				Observer:         observer,
				ConsumerName:     cfg.Idempotency.ConsumerName,
			},
		),
	}
}

func (w *ConsumerWorker) Run(ctx context.Context) {
	if !w.enabled {
		w.log.Info("Onboarding consumer worker disabled")
		return
	}

	defer func() {
		if err := w.runner.Close(); err != nil {
			w.log.Error("Close onboarding consumer failed", "error", err)
		}
	}()

	for ctx.Err() == nil {
		if err := w.consumer.Run(ctx); err != nil && ctx.Err() == nil {
			w.log.Error("Onboarding consumer failed", "error", err)
		}
	}
}

func NewConsumerGroupRunner(
	factory pdkafka.ConsumerGroupFactory,
	cfg *pdkafka.Config,
) (pdkafka.ConsumerGroupRunner, error) {
	groupID := cfg.ConsumerGroupPrefix + ".consul-bridge"
	group, err := factory.New(groupID)
	if err != nil {
		return nil, fmt.Errorf("create onboarding consumer group: %w", err)
	}
	return pdkafka.NewConsumerGroupRunner(group), nil
}
