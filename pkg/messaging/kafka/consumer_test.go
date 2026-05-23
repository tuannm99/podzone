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

	err = cgh.ConsumeClaim(session, claim)
	require.Error(t, err)
	assert.Equal(t, 0, session.marked)
}
