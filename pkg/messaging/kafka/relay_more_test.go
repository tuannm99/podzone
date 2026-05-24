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

type markPublishedFailStore struct {
	fakeOutboxStore
	markPublishedErr error
}

func (f *markPublishedFailStore) MarkPublished(ctx context.Context, ids []string, publishedAt time.Time) error {
	_ = ctx
	_ = ids
	_ = publishedAt
	return f.markPublishedErr
}

func TestNewRelayDefaultsLimit(t *testing.T) {
	relay := NewRelay(&fakeOutboxStore{}, &fakePublisher{}, 0)
	assert.Equal(t, 100, relay.limit)
}

func TestRelayRunOnceReturnsListError(t *testing.T) {
	store := &fakeOutboxStore{listErr: errors.New("list failed")}
	relay := NewRelay(store, &fakePublisher{}, 10)

	err := relay.RunOnce(context.Background())
	require.Error(t, err)
	assert.EqualError(t, err, "list failed")
}

func TestRelayRunOnceReturnsMarkPublishedError(t *testing.T) {
	store := &markPublishedFailStore{
		fakeOutboxStore: fakeOutboxStore{
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
			},
		},
		markPublishedErr: errors.New("mark published failed"),
	}
	relay := NewRelay(store, &fakePublisher{}, 10)

	err := relay.RunOnce(context.Background())
	require.Error(t, err)
	assert.EqualError(t, err, "mark published failed")
}
