package interactor

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/tuannm99/podzone/pkg/messaging"
)

func (s *interactor) appendOutboxRecord(ctx context.Context, now time.Time, record messaging.OutboxRecord) error {
	if s.outbox == nil {
		return nil
	}
	record.Status = "pending"
	record.Attempts = 0
	record.NextAttemptAt = now
	record.CreatedAt = now
	record.UpdatedAt = now
	return s.outbox.Append(ctx, nil, record)
}

func newIAMEventOutboxRecord(
	now time.Time,
	eventType string,
	tenantID string,
	entityID string,
	messageKey string,
	payload map[string]any,
) (messaging.OutboxRecord, error) {
	rawPayload, err := json.Marshal(payload)
	if err != nil {
		return messaging.OutboxRecord{}, err
	}

	if messageKey == "" {
		messageKey = entityID
	}

	return messaging.OutboxRecord{
		ID:         uuid.NewString(),
		Topic:      messaging.Topic("iam", "events"),
		MessageKey: messageKey,
		Envelope: messaging.Envelope{
			ID:            uuid.NewString(),
			Type:          eventType,
			Source:        "iam",
			TenantID:      tenantID,
			EntityID:      entityID,
			OccurredAt:    now,
			SchemaVersion: 1,
			Payload:       rawPayload,
		},
	}, nil
}
