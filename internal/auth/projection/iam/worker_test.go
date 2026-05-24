package iam

import (
	"context"
	"testing"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tuannm99/podzone/pkg/pdkafka"
)

type fakeConsumerGroupFactory struct {
	groupID string
	err     error
}

func (f *fakeConsumerGroupFactory) New(groupID string) (sarama.ConsumerGroup, error) {
	f.groupID = groupID
	if f.err != nil {
		return nil, f.err
	}
	return &fakeSaramaConsumerGroup{}, nil
}

type fakeSaramaConsumerGroup struct{}

func (f *fakeSaramaConsumerGroup) Consume(ctx context.Context, topics []string, handler sarama.ConsumerGroupHandler) error {
	return nil
}
func (f *fakeSaramaConsumerGroup) Errors() <-chan error      { return nil }
func (f *fakeSaramaConsumerGroup) Pause(map[string][]int32)  {}
func (f *fakeSaramaConsumerGroup) Resume(map[string][]int32) {}
func (f *fakeSaramaConsumerGroup) PauseAll()                 {}
func (f *fakeSaramaConsumerGroup) ResumeAll()                {}
func (f *fakeSaramaConsumerGroup) Close() error              { return nil }

func TestNewConsumerGroupRunner(t *testing.T) {
	factory := &fakeConsumerGroupFactory{}
	runner, err := NewConsumerGroupRunner(factory, &pdkafka.Config{ConsumerGroupPrefix: "podzone.auth"})
	require.NoError(t, err)
	require.NotNil(t, runner)
	assert.Equal(t, "podzone.auth.iam-projection", factory.groupID)
}
