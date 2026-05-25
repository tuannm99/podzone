package messaging

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeRedrivePublisher struct {
	requests []PublishRequest
}

func (f *fakeRedrivePublisher) Publish(ctx context.Context, topic string, key string, msg Envelope) error {
	f.requests = append(f.requests, PublishRequest{Topic: topic, Key: key, Msg: msg})
	return nil
}

func (f *fakeRedrivePublisher) PublishBatch(ctx context.Context, topic string, msgs []PublishRequest) error {
	f.requests = append(f.requests, msgs...)
	return nil
}

func TestRedriverRedrive(t *testing.T) {
	publisher := &fakeRedrivePublisher{}
	redriver := NewRedriver(publisher)
	env := Envelope{
		ID:            "evt_1",
		Type:          "tenant.created",
		Source:        "iam",
		OccurredAt:    time.Now().UTC(),
		SchemaVersion: 1,
		Payload:       json.RawMessage(`{"tenant_id":"t1"}`),
	}

	err := redriver.Redrive(context.Background(), RedriveRequest{
		SourceTopic: "podzone.iam.events.dlt",
		TargetTopic: "podzone.iam.events",
		Key:         "t1",
		Envelope:    env,
	})
	require.NoError(t, err)
	require.Len(t, publisher.requests, 1)
	assert.Equal(t, "podzone.iam.events", publisher.requests[0].Topic)
	assert.Equal(t, 1, ReadDeliveryMetadata(publisher.requests[0].Msg).RedriveCount)
}
