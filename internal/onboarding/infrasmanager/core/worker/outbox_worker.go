package worker

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tuannm99/podzone/internal/onboarding/infrasmanager/core"
	"github.com/tuannm99/podzone/internal/onboarding/infrasmanager/core/publisher"
	"github.com/tuannm99/podzone/pkg/pdlog"
)

// OutboxWorker polls Mongo outbox and publishes to Consul.
type OutboxWorker struct {
	log pdlog.Logger
	st  core.ConnectionStore
	pub *publisher.ConsulPublisher

	interval time.Duration
}

func NewOutboxWorker(log pdlog.Logger, st core.ConnectionStore, pub *publisher.ConsulPublisher) *OutboxWorker {
	return &OutboxWorker{
		log:      log,
		st:       st,
		pub:      pub,
		interval: 2 * time.Second,
	}
}

func (w *OutboxWorker) Run(ctx context.Context) {
	t := time.NewTicker(w.interval)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			w.tick(ctx)
		}
	}
}

func (w *OutboxWorker) tick(ctx context.Context) {
	msgs, err := w.st.FindDueOutbox(50)
	if err != nil {
		w.log.Warn("outbox poll failed", "error", err)
		return
	}

	for _, m := range msgs {
		switch m.Topic {
		case "consul.publish":
			key, _ := m.Payload["key"].(string)
			val, _ := m.Payload["value"].(string)

			if err := w.pub.Put(ctx, key, val); err != nil {
				next := time.Now().Add(publisher.DefaultBackoff(m.RetryCount + 1))
				_ = w.st.MarkOutboxFailed(m.EventID, next)

				_ = w.st.AppendEvent(core.ConnectionEvent{
					ID:            uuid.NewString(),
					CorrelationID: m.CorrelationID,
					TenantID:      m.TenantID,
					InfraType:     m.InfraType,
					Name:          m.Name,
					Action:        "publish_consul",
					Status:        "failed",
					Error:         err.Error(),
					Result:        map[string]interface{}{"key": key, "next_retry": next.UTC().Format(time.RFC3339)},
					CreatedAt:     time.Now(),
				})
				continue
			}

			_ = w.st.MarkOutboxDone(m.EventID)
			_ = w.st.AppendEvent(core.ConnectionEvent{
				ID:            uuid.NewString(),
				CorrelationID: m.CorrelationID,
				TenantID:      m.TenantID,
				InfraType:     m.InfraType,
				Name:          m.Name,
				Action:        "publish_consul",
				Status:        "succeeded",
				Result:        map[string]interface{}{"key": key, "bytes": len(val)},
				CreatedAt:     time.Now(),
			})

		case "consul.delete":
			key, _ := m.Payload["key"].(string)
			if err := w.pub.Delete(ctx, key); err != nil {
				next := time.Now().Add(publisher.DefaultBackoff(m.RetryCount + 1))
				_ = w.st.MarkOutboxFailed(m.EventID, next)

				_ = w.st.AppendEvent(core.ConnectionEvent{
					ID:            uuid.NewString(),
					CorrelationID: m.CorrelationID,
					TenantID:      m.TenantID,
					InfraType:     m.InfraType,
					Name:          m.Name,
					Action:        "delete_consul",
					Status:        "failed",
					Error:         err.Error(),
					Result:        map[string]interface{}{"key": key, "next_retry": next.UTC().Format(time.RFC3339)},
					CreatedAt:     time.Now(),
				})
				continue
			}

			_ = w.st.MarkOutboxDone(m.EventID)
			_ = w.st.AppendEvent(core.ConnectionEvent{
				ID:            uuid.NewString(),
				CorrelationID: m.CorrelationID,
				TenantID:      m.TenantID,
				InfraType:     m.InfraType,
				Name:          m.Name,
				Action:        "delete_consul",
				Status:        "succeeded",
				Result:        map[string]interface{}{"key": key},
				CreatedAt:     time.Now(),
			})
		}
	}
}
