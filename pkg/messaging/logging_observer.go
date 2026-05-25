package messaging

import (
	"context"

	"github.com/tuannm99/podzone/pkg/pdlog"
)

type LoggingObserver struct {
	log       pdlog.Logger
	component string
	cfg       ObservabilityConfig
}

func NewLoggingObserver(log pdlog.Logger, component string, cfg ConsumerRuntimeConfig) Observer {
	if !cfg.Observability.Enabled {
		return ObserverFunc(func(context.Context, DeliveryEvent) {})
	}
	return &LoggingObserver{
		log:       log,
		component: component,
		cfg:       cfg.Observability,
	}
}

func (o *LoggingObserver) Observe(_ context.Context, event DeliveryEvent) {
	fields := []any{
		"component", o.component,
		"consumer", event.ConsumerName,
		"topic", event.Topic,
		"message_id", event.Envelope.ID,
		"reason", event.Reason,
		"error", event.Err,
	}

	switch event.Action {
	case FailureActionRetry:
		if o.cfg.LogRetries {
			o.log.Warn("Messaging consumer retry", fields...)
		}
	case FailureActionDeadLetter:
		if o.cfg.LogDeadLetters {
			o.log.Error("Messaging consumer dead-lettered", fields...)
		}
	case FailureActionDrop:
		if o.cfg.LogDrops {
			o.log.Warn("Messaging consumer dropped event", fields...)
		}
	default:
		if o.cfg.LogHandled {
			o.log.Info("Messaging consumer handled event", fields[:8]...)
		}
	}
}
