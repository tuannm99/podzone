package iam

import (
	"context"
	"fmt"

	messagingkafka "github.com/tuannm99/podzone/pkg/messaging/kafka"
	"github.com/tuannm99/podzone/pkg/pdkafka"
	"github.com/tuannm99/podzone/pkg/pdlog"
)

type Worker struct {
	log      pdlog.Logger
	runner   pdkafka.ConsumerGroupRunner
	consumer *messagingkafka.Consumer
}

func NewWorker(log pdlog.Logger, runner pdkafka.ConsumerGroupRunner, handler *Handler) *Worker {
	return &Worker{
		log:      log,
		runner:   runner,
		consumer: messagingkafka.NewConsumer(runner, []string{"podzone.iam.events"}, handler),
	}
}

func (w *Worker) Run(ctx context.Context) {
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
