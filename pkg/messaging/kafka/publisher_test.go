package kafka

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/IBM/sarama"
	saramamocks "github.com/IBM/sarama/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tuannm99/podzone/pkg/messaging"
)

func validEnvelope() messaging.Envelope {
	return messaging.Envelope{
		ID:            "evt_1",
		Type:          "tenant.created",
		Source:        "iam",
		OccurredAt:    time.Now().UTC(),
		SchemaVersion: 1,
		Headers: map[string]string{
			"tenant_id": "t1",
		},
		Payload: json.RawMessage(`{"tenant_id":"t1"}`),
	}
}

func TestPublisherPublish(t *testing.T) {
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	producer := saramamocks.NewSyncProducer(t, cfg)
	producer.ExpectSendMessageAndSucceed()

	pub := NewPublisher(producer)
	err := pub.Publish(context.Background(), "podzone.iam.events", "t1", validEnvelope())
	require.NoError(t, err)
}

func TestPublisherPublishBatch(t *testing.T) {
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	producer := saramamocks.NewSyncProducer(t, cfg)
	producer.ExpectSendMessageAndSucceed()
	producer.ExpectSendMessageAndSucceed()

	pub := NewPublisher(producer)
	err := pub.PublishBatch(context.Background(), "podzone.iam.events", []messaging.PublishRequest{
		{Key: "t1", Msg: validEnvelope()},
		{Topic: "podzone.iam.audit", Key: "t2", Msg: validEnvelope()},
	})
	require.NoError(t, err)
}

func TestPublisherRejectsInvalidEnvelope(t *testing.T) {
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	producer := saramamocks.NewSyncProducer(t, cfg)

	pub := NewPublisher(producer)
	err := pub.Publish(context.Background(), "podzone.iam.events", "t1", messaging.Envelope{})
	require.Error(t, err)
	assert.ErrorIs(t, err, messaging.ErrEmptyEnvelopeID)
}
