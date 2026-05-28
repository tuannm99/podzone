package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/tuannm99/podzone/pkg/messaging"
)

const onboardingEventsTopic = "podzone.onboarding.events"

type OutboxStoreAdapter struct {
	store ConnectionStore
}

var _ messaging.OutboxStore = (*OutboxStoreAdapter)(nil)

func NewOutboxStoreAdapter(store ConnectionStore) *OutboxStoreAdapter {
	return &OutboxStoreAdapter{store: store}
}

func (a *OutboxStoreAdapter) Append(ctx context.Context, tx messaging.Tx, record messaging.OutboxRecord) error {
	_ = ctx
	_ = tx
	if a == nil || a.store == nil {
		return errors.New("onboarding outbox adapter: nil store")
	}
	payload, err := outboxPayloadToMap(record.Envelope.Payload)
	if err != nil {
		return err
	}
	msg := OutboxMessage{
		EventID:       record.ID,
		CorrelationID: record.Envelope.CorrelationID,
		Topic:         record.Envelope.Type,
		Payload:       payload,
		TenantID:      record.Envelope.TenantID,
		InfraType:     InfraType(record.Envelope.Headers["infra_type"]),
		Name:          record.Envelope.EntityID,
		Status:        record.Status,
		RetryCount:    record.Attempts,
		NextRetry:     record.NextAttemptAt,
		CreatedAt:     record.CreatedAt,
		UpdatedAt:     record.UpdatedAt,
	}
	if msg.Name == "" {
		msg.Name = "default"
	}
	if record.Envelope.Headers != nil {
		if tenantID := record.Envelope.Headers["tenant_id"]; tenantID != "" {
			msg.TenantID = tenantID
		}
	}
	return a.store.EnqueueOutbox(msg)
}

func (a *OutboxStoreAdapter) ListPending(ctx context.Context, limit int) ([]messaging.OutboxRecord, error) {
	_ = ctx
	if a == nil || a.store == nil {
		return nil, errors.New("onboarding outbox adapter: nil store")
	}
	msgs, err := a.store.FindDueOutbox(limit)
	if err != nil {
		return nil, err
	}
	out := make([]messaging.OutboxRecord, 0, len(msgs))
	for _, msg := range msgs {
		payload, err := json.Marshal(msg.Payload)
		if err != nil {
			return nil, fmt.Errorf("marshal onboarding outbox payload: %w", err)
		}
		out = append(out, messaging.OutboxRecord{
			ID:         msg.EventID,
			Topic:      onboardingEventsTopic,
			MessageKey: onboardingOutboxKey(msg),
			Envelope: messaging.Envelope{
				ID:            msg.EventID,
				Type:          msg.Topic,
				Source:        "onboarding",
				TenantID:      msg.TenantID,
				EntityID:      msg.Name,
				CorrelationID: msg.CorrelationID,
				OccurredAt:    msg.CreatedAt,
				SchemaVersion: 1,
				Headers: map[string]string{
					"infra_type": string(msg.InfraType),
				},
				Payload: payload,
			},
			Status:        msg.Status,
			Attempts:      msg.RetryCount,
			NextAttemptAt: msg.NextRetry,
			CreatedAt:     msg.CreatedAt,
			UpdatedAt:     msg.UpdatedAt,
		})
	}
	return out, nil
}

func (a *OutboxStoreAdapter) MarkPublished(ctx context.Context, ids []string, publishedAt time.Time) error {
	_ = ctx
	if a == nil || a.store == nil {
		return errors.New("onboarding outbox adapter: nil store")
	}
	for _, id := range ids {
		if err := a.store.MarkOutboxDone(id); err != nil {
			return err
		}
	}
	return nil
}

func (a *OutboxStoreAdapter) MarkFailed(ctx context.Context, id string, errText string, nextAttemptAt time.Time) error {
	_ = ctx
	if a == nil || a.store == nil {
		return errors.New("onboarding outbox adapter: nil store")
	}
	return a.store.MarkOutboxFailed(id, nextAttemptAt)
}

func onboardingOutboxKey(msg OutboxMessage) string {
	if msg.TenantID == "" {
		return msg.Name
	}
	if msg.Name == "" {
		return msg.TenantID
	}
	return fmt.Sprintf("%s:%s:%s", msg.TenantID, msg.InfraType, msg.Name)
}

func outboxPayloadToMap(payload []byte) (map[string]interface{}, error) {
	if len(payload) == 0 {
		return map[string]interface{}{}, nil
	}
	var out map[string]interface{}
	if err := json.Unmarshal(payload, &out); err != nil {
		return nil, fmt.Errorf("unmarshal onboarding outbox payload: %w", err)
	}
	return out, nil
}
