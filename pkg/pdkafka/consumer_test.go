package pdkafka

import (
	"context"
	"errors"
	"testing"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeConsumerGroup struct {
	consumeTopics []string
	consumeErr    error
	closeErr      error
}

func (f *fakeConsumerGroup) Consume(
	ctx context.Context,
	topics []string,
	handler sarama.ConsumerGroupHandler,
) error {
	_ = ctx
	_ = handler
	f.consumeTopics = append([]string(nil), topics...)
	return f.consumeErr
}

func (f *fakeConsumerGroup) Errors() <-chan error { return nil }
func (f *fakeConsumerGroup) Close() error         { return f.closeErr }
func (f *fakeConsumerGroup) Pause(map[string][]int32) {
}

func (f *fakeConsumerGroup) Resume(map[string][]int32) {
}

func (f *fakeConsumerGroup) PauseAll() {
}

func (f *fakeConsumerGroup) ResumeAll() {
}

type noopConsumerGroupHandler struct{}

func (noopConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (noopConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }
func (noopConsumerGroupHandler) ConsumeClaim(sarama.ConsumerGroupSession, sarama.ConsumerGroupClaim) error {
	return nil
}

func TestConsumerGroupRunnerRunDelegates(t *testing.T) {
	group := &fakeConsumerGroup{}
	runner := NewConsumerGroupRunner(group)

	err := runner.Run(context.Background(), []string{"podzone.iam.events"}, noopConsumerGroupHandler{})
	require.NoError(t, err)
	assert.Equal(t, []string{"podzone.iam.events"}, group.consumeTopics)
}

func TestConsumerGroupRunnerCloseDelegates(t *testing.T) {
	group := &fakeConsumerGroup{closeErr: errors.New("close failed")}
	runner := NewConsumerGroupRunner(group)

	err := runner.Close()
	require.Error(t, err)
	assert.EqualError(t, err, "close failed")
}
