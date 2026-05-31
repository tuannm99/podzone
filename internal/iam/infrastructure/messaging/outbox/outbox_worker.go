package outbox

import (
	"context"
	"errors"
	"time"

	"github.com/tuannm99/podzone/pkg/messaging"
	messagingkafka "github.com/tuannm99/podzone/pkg/messaging/kafka"
	"github.com/tuannm99/podzone/pkg/pdlog"
)

type OutboxWorker struct {
	log      pdlog.Logger
	relay    *messagingkafka.Relay
	interval time.Duration
}

func NewOutboxWorker(log pdlog.Logger, relay *messagingkafka.Relay) *OutboxWorker {
	return &OutboxWorker{
		log:      log,
		relay:    relay,
		interval: 5 * time.Second,
	}
}

func (w *OutboxWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.tick(ctx)
		}
	}
}

func (w *OutboxWorker) tick(ctx context.Context) {
	if err := w.relay.RunOnce(ctx); err != nil {
		if errors.Is(err, messaging.ErrNoMessages) {
			return
		}
		w.log.Error("IAM outbox relay tick failed", "error", err)
	}
}
