package messaging

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeTxOutboxStore struct {
	record OutboxRecord
}

func (f *fakeTxOutboxStore) Append(ctx context.Context, tx Tx, record OutboxRecord) error {
	f.record = record
	return nil
}

func (f *fakeTxOutboxStore) ListPending(ctx context.Context, limit int) ([]OutboxRecord, error) {
	return nil, nil
}

func (f *fakeTxOutboxStore) MarkPublished(ctx context.Context, ids []string, publishedAt time.Time) error {
	return nil
}

func (f *fakeTxOutboxStore) MarkFailed(ctx context.Context, id string, errText string, nextAttemptAt time.Time) error {
	return nil
}

func TestTransactionalOutboxPublisherPublish(t *testing.T) {
	store := &fakeTxOutboxStore{}
	now := time.Date(2026, 5, 25, 12, 0, 0, 0, time.UTC)
	publisher := NewTransactionalOutboxPublisher(store, func() time.Time { return now })
	msg := Envelope{
		ID:            "evt_1",
		Type:          "tenant.created",
		Source:        "iam",
		OccurredAt:    now,
		SchemaVersion: 1,
		Payload:       json.RawMessage(`{"tenant_id":"t1"}`),
	}

	err := publisher.Publish(context.Background(), nil, "podzone.iam.events", "t1", msg)
	require.NoError(t, err)
	assert.Equal(t, "evt_1", store.record.ID)
	assert.Equal(t, "podzone.iam.events", store.record.Topic)
	assert.Equal(t, "pending", store.record.Status)
}
