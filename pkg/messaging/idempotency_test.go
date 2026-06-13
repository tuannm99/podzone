package messaging

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeInboxStore struct {
	beginDecision InboxDecision
	beginErr      error
	completed     []string
	failed        []string
}

func (f *fakeInboxStore) Begin(
	ctx context.Context,
	consumerName string,
	messageID string,
	now time.Time,
) (InboxDecision, error) {
	return f.beginDecision, f.beginErr
}

func (f *fakeInboxStore) Complete(
	ctx context.Context,
	consumerName string,
	messageID string,
	processedAt time.Time,
) error {
	f.completed = append(f.completed, consumerName+":"+messageID)
	return nil
}

func (f *fakeInboxStore) Fail(
	ctx context.Context,
	consumerName string,
	messageID string,
	errText string,
	failedAt time.Time,
) error {
	f.failed = append(f.failed, consumerName+":"+messageID)
	return nil
}

func TestIdempotentConsumerDuplicate(t *testing.T) {
	store := &fakeInboxStore{beginDecision: InboxDecisionDuplicate}
	called := false
	handler := HandlerFunc(func(ctx context.Context, msg Envelope) error {
		called = true
		return nil
	})

	err := IdempotentConsumer(store, "proj", nil)(handler).Handle(context.Background(), Envelope{ID: "evt_1"})
	require.NoError(t, err)
	assert.False(t, called)
}

func TestIdempotentConsumerCompleteAndFail(t *testing.T) {
	store := &fakeInboxStore{beginDecision: InboxDecisionAcquired}
	okHandler := HandlerFunc(func(ctx context.Context, msg Envelope) error { return nil })
	err := IdempotentConsumer(store, "proj", nil)(okHandler).Handle(context.Background(), Envelope{ID: "evt_1"})
	require.NoError(t, err)
	assert.Equal(t, []string{"proj:evt_1"}, store.completed)

	store = &fakeInboxStore{beginDecision: InboxDecisionAcquired}
	failHandler := HandlerFunc(func(ctx context.Context, msg Envelope) error { return errors.New("boom") })
	err = IdempotentConsumer(store, "proj", nil)(failHandler).Handle(context.Background(), Envelope{ID: "evt_2"})
	require.Error(t, err)
	assert.Equal(t, []string{"proj:evt_2"}, store.failed)
}
