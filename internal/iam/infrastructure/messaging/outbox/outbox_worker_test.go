package outbox

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tuannm99/podzone/pkg/messaging"
	messagingkafka "github.com/tuannm99/podzone/pkg/messaging/kafka"
	"github.com/tuannm99/podzone/pkg/pdlog"
)

type fakeOutboxStore struct {
	listPending func(ctx context.Context, limit int) ([]messaging.OutboxRecord, error)
	markPub     func(ctx context.Context, ids []string, publishedAt time.Time) error
	markFail    func(ctx context.Context, id string, errText string, nextAttemptAt time.Time) error
}

func (f *fakeOutboxStore) Append(ctx context.Context, tx messaging.Tx, record messaging.OutboxRecord) error {
	return nil
}

func (f *fakeOutboxStore) ListPending(ctx context.Context, limit int) ([]messaging.OutboxRecord, error) {
	if f.listPending != nil {
		return f.listPending(ctx, limit)
	}
	return nil, nil
}

func (f *fakeOutboxStore) MarkPublished(ctx context.Context, ids []string, publishedAt time.Time) error {
	if f.markPub != nil {
		return f.markPub(ctx, ids, publishedAt)
	}
	return nil
}

func (f *fakeOutboxStore) MarkFailed(ctx context.Context, id string, errText string, nextAttemptAt time.Time) error {
	if f.markFail != nil {
		return f.markFail(ctx, id, errText, nextAttemptAt)
	}
	return nil
}

type fakePublisher struct {
	publish func(ctx context.Context, topic string, key string, msg messaging.Envelope) error
}

func (f *fakePublisher) Publish(ctx context.Context, topic string, key string, msg messaging.Envelope) error {
	if f.publish != nil {
		return f.publish(ctx, topic, key, msg)
	}
	return nil
}

func (f *fakePublisher) PublishBatch(ctx context.Context, topic string, msgs []messaging.PublishRequest) error {
	return nil
}

func TestOutboxWorkerTick_IgnoresNoMessages(t *testing.T) {
	relay := messagingkafka.NewRelay(&fakeOutboxStore{}, &fakePublisher{}, 10)
	worker := NewOutboxWorker(pdlog.NopLogger{}, relay)

	require.NotPanics(t, func() {
		worker.tick(context.Background())
	})
}

func TestOutboxWorkerTick_RunsRelay(t *testing.T) {
	called := false
	relay := messagingkafka.NewRelay(&fakeOutboxStore{
		listPending: func(ctx context.Context, limit int) ([]messaging.OutboxRecord, error) {
			return []messaging.OutboxRecord{{
				ID:    "m1",
				Topic: "podzone.iam.events",
				Envelope: messaging.Envelope{
					ID:            "e1",
					Type:          "tenant.created",
					Source:        "iam",
					OccurredAt:    time.Now().UTC(),
					SchemaVersion: 1,
					Payload:       []byte(`{"tenant_id":"t1"}`),
				},
			}}, nil
		},
		markPub: func(ctx context.Context, ids []string, publishedAt time.Time) error {
			called = true
			assert.Equal(t, []string{"m1"}, ids)
			return nil
		},
	}, &fakePublisher{
		publish: func(ctx context.Context, topic string, key string, msg messaging.Envelope) error {
			assert.Equal(t, "podzone.iam.events", topic)
			return nil
		},
	}, 10)
	worker := NewOutboxWorker(pdlog.NopLogger{}, relay)

	worker.tick(context.Background())

	assert.True(t, called)
}

func TestOutboxWorkerTick_SwallowsRelayFailures(t *testing.T) {
	expectedErr := errors.New("boom")
	relay := messagingkafka.NewRelay(&fakeOutboxStore{
		listPending: func(ctx context.Context, limit int) ([]messaging.OutboxRecord, error) {
			return nil, expectedErr
		},
	}, &fakePublisher{}, 10)
	worker := NewOutboxWorker(pdlog.NopLogger{}, relay)

	require.NotPanics(t, func() {
		worker.tick(context.Background())
	})
}
