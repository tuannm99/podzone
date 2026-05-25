package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tuannm99/podzone/pkg/messaging"
)

func validConsumerEnvelope() messaging.Envelope {
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

type fakeRunner struct {
	topics  []string
	handler sarama.ConsumerGroupHandler
	runErr  error
}

func (f *fakeRunner) Run(ctx context.Context, topics []string, handler sarama.ConsumerGroupHandler) error {
	f.topics = append([]string(nil), topics...)
	f.handler = handler
	return f.runErr
}

func (f *fakeRunner) Close() error { return nil }

type fakeHandler struct {
	handled []messaging.Envelope
	err     error
}

func (f *fakeHandler) Handle(ctx context.Context, msg messaging.Envelope) error {
	f.handled = append(f.handled, msg)
	return f.err
}

type fakeConsumerPublisher struct {
	requests []messaging.PublishRequest
	err      error
}

func (f *fakeConsumerPublisher) Publish(ctx context.Context, topic string, key string, msg messaging.Envelope) error {
	if f.err != nil {
		return f.err
	}
	f.requests = append(f.requests, messaging.PublishRequest{Topic: topic, Key: key, Msg: msg})
	return nil
}

func (f *fakeConsumerPublisher) PublishBatch(ctx context.Context, topic string, msgs []messaging.PublishRequest) error {
	if f.err != nil {
		return f.err
	}
	f.requests = append(f.requests, msgs...)
	return nil
}

type fakeSession struct {
	ctx    context.Context
	marked int
}

func (s *fakeSession) Claims() map[string][]int32                  { return nil }
func (s *fakeSession) MemberID() string                            { return "" }
func (s *fakeSession) GenerationID() int32                         { return 0 }
func (s *fakeSession) MarkOffset(string, int32, int64, string)     {}
func (s *fakeSession) ResetOffset(string, int32, int64, string)    {}
func (s *fakeSession) MarkMessage(*sarama.ConsumerMessage, string) { s.marked++ }
func (s *fakeSession) Context() context.Context                    { return s.ctx }
func (s *fakeSession) Commit()                                     {}

type fakeClaim struct {
	messages chan *sarama.ConsumerMessage
}

func (c *fakeClaim) Topic() string              { return "podzone.iam.events" }
func (c *fakeClaim) Partition() int32           { return 0 }
func (c *fakeClaim) InitialOffset() int64       { return 0 }
func (c *fakeClaim) HighWaterMarkOffset() int64 { return 1 }
func (c *fakeClaim) Messages() <-chan *sarama.ConsumerMessage {
	return c.messages
}

func TestConsumerRunDelegatesToRunner(t *testing.T) {
	runner := &fakeRunner{}
	handler := &fakeHandler{}

	consumer := NewConsumer(runner, []string{"podzone.iam.events"}, handler)
	require.NoError(t, consumer.Run(context.Background()))
	assert.Equal(t, []string{"podzone.iam.events"}, runner.topics)
	require.NotNil(t, runner.handler)
}

func TestConsumerGroupHandlerConsumeClaim(t *testing.T) {
	handler := &fakeHandler{}
	cgh := &consumerGroupHandler{handler: handler}

	payload, err := json.Marshal(messaging.Envelope{
		ID:            "evt_1",
		Type:          "tenant.created",
		Source:        "iam",
		OccurredAt:    time.Now().UTC(),
		SchemaVersion: 1,
		Payload:       json.RawMessage(`{"tenant_id":"t1"}`),
	})
	require.NoError(t, err)

	msgCh := make(chan *sarama.ConsumerMessage, 1)
	msgCh <- &sarama.ConsumerMessage{Value: payload}
	close(msgCh)

	session := &fakeSession{ctx: context.Background()}
	claim := &fakeClaim{messages: msgCh}

	require.NoError(t, cgh.ConsumeClaim(session, claim))
	require.Len(t, handler.handled, 1)
	assert.Equal(t, 1, session.marked)
}

func TestConsumerGroupHandlerConsumeClaim_HandlerError(t *testing.T) {
	handler := &fakeHandler{err: errors.New("boom")}
	cgh := &consumerGroupHandler{handler: handler, opts: ConsumerOptions{}}

	payload, err := json.Marshal(messaging.Envelope{
		ID:            "evt_1",
		Type:          "tenant.created",
		Source:        "iam",
		OccurredAt:    time.Now().UTC(),
		SchemaVersion: 1,
		Payload:       json.RawMessage(`{"tenant_id":"t1"}`),
	})
	require.NoError(t, err)

	msgCh := make(chan *sarama.ConsumerMessage, 1)
	msgCh <- &sarama.ConsumerMessage{Value: payload}
	close(msgCh)

	session := &fakeSession{ctx: context.Background()}
	claim := &fakeClaim{messages: msgCh}

	err = cgh.ConsumeClaim(session, claim)
	require.Error(t, err)
	assert.Equal(t, 0, session.marked)
}

func TestConsumerGroupHandlerConsumeClaim_RetryablePublishesRetry(t *testing.T) {
	handler := &fakeHandler{err: messaging.RetryableError(errors.New("transient"), "retry later")}
	publisher := &fakeConsumerPublisher{}
	cgh := &consumerGroupHandler{
		handler: handler,
		opts: ConsumerOptions{
			Publisher:     publisher,
			RetryPolicy:   messaging.RetryPolicy{MaxAttempts: 3, BaseDelay: time.Second},
			Classifier:    messaging.DefaultErrorClassifier(),
			TopicStrategy: messaging.DefaultTopicStrategy(),
			ConsumerName:  "auth.iam-projection",
		},
	}

	payload, err := json.Marshal(validConsumerEnvelope())
	require.NoError(t, err)

	msgCh := make(chan *sarama.ConsumerMessage, 1)
	msgCh <- &sarama.ConsumerMessage{
		Topic: "podzone.iam.events",
		Key:   []byte("t1"),
		Value: payload,
	}
	close(msgCh)

	session := &fakeSession{ctx: context.Background()}
	claim := &fakeClaim{messages: msgCh}

	require.NoError(t, cgh.ConsumeClaim(session, claim))
	require.Len(t, publisher.requests, 1)
	assert.Equal(t, "podzone.iam.events.retry.1", publisher.requests[0].Topic)
	assert.Equal(t, 1, messaging.ReadDeliveryMetadata(publisher.requests[0].Msg).Attempt)
	assert.Equal(t, 1, session.marked)
}

func TestConsumerGroupHandlerConsumeClaim_RetryUsesOriginalTopic(t *testing.T) {
	handler := &fakeHandler{err: messaging.RetryableError(errors.New("transient"), "retry later")}
	publisher := &fakeConsumerPublisher{}
	env := validConsumerEnvelope()
	env = messaging.WithDeliveryMetadata(env, messaging.DeliveryMetadata{
		Attempt:       1,
		MaxAttempts:   5,
		OriginalTopic: "podzone.iam.events",
	})
	cgh := &consumerGroupHandler{
		handler: handler,
		opts: ConsumerOptions{
			Publisher:     publisher,
			RetryPolicy:   messaging.RetryPolicy{MaxAttempts: 5, BaseDelay: time.Second},
			Classifier:    messaging.DefaultErrorClassifier(),
			TopicStrategy: messaging.DefaultTopicStrategy(),
			ConsumerName:  "auth.iam-projection",
		},
	}

	payload, err := json.Marshal(env)
	require.NoError(t, err)

	msgCh := make(chan *sarama.ConsumerMessage, 1)
	msgCh <- &sarama.ConsumerMessage{
		Topic: "podzone.iam.events.retry.1",
		Key:   []byte("t1"),
		Value: payload,
	}
	close(msgCh)

	session := &fakeSession{ctx: context.Background()}
	claim := &fakeClaim{messages: msgCh}

	require.NoError(t, cgh.ConsumeClaim(session, claim))
	require.Len(t, publisher.requests, 1)
	assert.Equal(t, "podzone.iam.events.retry.2", publisher.requests[0].Topic)
}

func TestConsumerGroupHandlerConsumeClaim_PermanentPublishesDeadLetter(t *testing.T) {
	handler := &fakeHandler{err: messaging.DeadLetterError(errors.New("invalid"), "bad payload")}
	publisher := &fakeConsumerPublisher{}
	cgh := &consumerGroupHandler{
		handler: handler,
		opts: ConsumerOptions{
			Publisher:        publisher,
			DeadLetterPolicy: messaging.DeadLetterPolicy{Strategy: messaging.DefaultTopicStrategy()},
			Classifier:       messaging.DefaultErrorClassifier(),
			TopicStrategy:    messaging.DefaultTopicStrategy(),
			ConsumerName:     "auth.iam-projection",
		},
	}

	payload, err := json.Marshal(validConsumerEnvelope())
	require.NoError(t, err)

	msgCh := make(chan *sarama.ConsumerMessage, 1)
	msgCh <- &sarama.ConsumerMessage{
		Topic: "podzone.iam.events",
		Key:   []byte("t1"),
		Value: payload,
	}
	close(msgCh)

	session := &fakeSession{ctx: context.Background()}
	claim := &fakeClaim{messages: msgCh}

	require.NoError(t, cgh.ConsumeClaim(session, claim))
	require.Len(t, publisher.requests, 1)
	assert.Equal(t, "podzone.iam.events.dlt", publisher.requests[0].Topic)
	assert.Equal(t, "bad payload", messaging.ReadDeliveryMetadata(publisher.requests[0].Msg).DeadLetterReason)
	assert.Equal(t, 1, session.marked)
}

func TestConsumerGroupHandlerConsumeClaim_DeadLetterUsesOriginalTopic(t *testing.T) {
	handler := &fakeHandler{err: messaging.DeadLetterError(errors.New("invalid"), "bad payload")}
	publisher := &fakeConsumerPublisher{}
	env := validConsumerEnvelope()
	env = messaging.WithDeliveryMetadata(env, messaging.DeliveryMetadata{
		Attempt:       2,
		MaxAttempts:   5,
		OriginalTopic: "podzone.iam.events",
	})
	cgh := &consumerGroupHandler{
		handler: handler,
		opts: ConsumerOptions{
			Publisher:        publisher,
			DeadLetterPolicy: messaging.DeadLetterPolicy{Strategy: messaging.DefaultTopicStrategy()},
			Classifier:       messaging.DefaultErrorClassifier(),
			TopicStrategy:    messaging.DefaultTopicStrategy(),
			ConsumerName:     "auth.iam-projection",
		},
	}

	payload, err := json.Marshal(env)
	require.NoError(t, err)

	msgCh := make(chan *sarama.ConsumerMessage, 1)
	msgCh <- &sarama.ConsumerMessage{
		Topic: "podzone.iam.events.retry.2",
		Key:   []byte("t1"),
		Value: payload,
	}
	close(msgCh)

	session := &fakeSession{ctx: context.Background()}
	claim := &fakeClaim{messages: msgCh}

	require.NoError(t, cgh.ConsumeClaim(session, claim))
	require.Len(t, publisher.requests, 1)
	assert.Equal(t, "podzone.iam.events.dlt", publisher.requests[0].Topic)
}

func TestConsumerGroupHandlerDecodeEnvelopeMergesRecordHeaders(t *testing.T) {
	cgh := &consumerGroupHandler{opts: ConsumerOptions{}}
	payload, err := json.Marshal(validConsumerEnvelope())
	require.NoError(t, err)

	env, err := cgh.decodeEnvelope(&sarama.ConsumerMessage{
		Value: payload,
		Headers: []*sarama.RecordHeader{
			{Key: []byte(messaging.HeaderAttempt), Value: []byte("2")},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, 2, messaging.ReadDeliveryMetadata(env).Attempt)
}
