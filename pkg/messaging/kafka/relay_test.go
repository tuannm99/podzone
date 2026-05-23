package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tuannm99/podzone/pkg/messaging"
)

type fakeOutboxStore struct {
	items           []messaging.OutboxRecord
	listErr         error
	markedPublished []string
	markedFailed    []string
}

func (f *fakeOutboxStore) Append(ctx context.Context, tx messaging.Tx, record messaging.OutboxRecord) error {
	return nil
}

func (f *fakeOutboxStore) ListPending(ctx context.Context, limit int) ([]messaging.OutboxRecord, error) {
	return f.items, f.listErr
}

func (f *fakeOutboxStore) MarkPublished(ctx context.Context, ids []string, publishedAt time.Time) error {
	f.markedPublished = append(f.markedPublished, ids...)
	return nil
}

func (f *fakeOutboxStore) MarkFailed(ctx context.Context, id string, errText string, nextAttemptAt time.Time) error {
	f.markedFailed = append(f.markedFailed, id)
	return nil
}

type fakePublisher struct {
	errFor map[string]error
}

func (f *fakePublisher) Publish(ctx context.Context, topic string, key string, msg messaging.Envelope) error {
	return f.errFor[msg.ID]
}

func (f *fakePublisher) PublishBatch(ctx context.Context, topic string, msgs []messaging.PublishRequest) error {
	return nil
}

func TestRelayRunOnce_NoMessages(t *testing.T) {
	store := &fakeOutboxStore{}
	relay := NewRelay(store, &fakePublisher{}, 10)

	err := relay.RunOnce(context.Background())
	require.ErrorIs(t, err, messaging.ErrNoMessages)
}

func TestRelayRunOnce_MarksPublishedAndFailed(t *testing.T) {
	store := &fakeOutboxStore{
		items: []messaging.OutboxRecord{
			{
				ID:         "1",
				Topic:      "podzone.iam.events",
				MessageKey: "t1",
				Envelope: messaging.Envelope{
					ID:            "evt_1",
					Type:          "tenant.created",
					Source:        "iam",
					OccurredAt:    time.Now().UTC(),
					SchemaVersion: 1,
					Payload:       json.RawMessage(`{"tenant_id":"t1"}`),
				},
			},
			{
				ID:         "2",
				Topic:      "podzone.iam.events",
				MessageKey: "t2",
				Envelope: messaging.Envelope{
					ID:            "evt_2",
					Type:          "tenant.created",
					Source:        "iam",
					OccurredAt:    time.Now().UTC(),
					SchemaVersion: 1,
					Payload:       json.RawMessage(`{"tenant_id":"t2"}`),
				},
			},
		},
	}
	publisher := &fakePublisher{
		errFor: map[string]error{
			"evt_2": errors.New("publish failed"),
		},
	}
	relay := NewRelay(store, publisher, 10)

	require.NoError(t, relay.RunOnce(context.Background()))
	assert.Equal(t, []string{"1"}, store.markedPublished)
	assert.Equal(t, []string{"2"}, store.markedFailed)
}
