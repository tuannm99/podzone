package pdkafka

import (
	"context"
	"errors"
	"testing"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"

	"github.com/tuannm99/podzone/pkg/pdlog"
)

type fakeProducer struct {
	closeErr error
	closed   bool
}

func (f *fakeProducer) SendMessage(msg *sarama.ProducerMessage) (int32, int64, error) {
	return 0, 0, nil
}
func (f *fakeProducer) SendMessages(msgs []*sarama.ProducerMessage) error { return nil }
func (f *fakeProducer) Close() error {
	f.closed = true
	return f.closeErr
}

type fakeAdmin struct {
	describeErr error
	createErr   error
	topics      map[string]sarama.TopicDetail
	closeErr    error
	closed      bool
}

func (f *fakeAdmin) DescribeCluster() ([]*sarama.Broker, int32, error) {
	return nil, 7, f.describeErr
}

func (f *fakeAdmin) CreateTopic(topic string, detail *sarama.TopicDetail, validateOnly bool) error {
	if f.createErr != nil {
		return f.createErr
	}
	if f.topics == nil {
		f.topics = map[string]sarama.TopicDetail{}
	}
	if detail != nil {
		f.topics[topic] = *detail
	}
	return nil
}

func (f *fakeAdmin) ListTopics() (map[string]sarama.TopicDetail, error) {
	if f.topics == nil {
		return map[string]sarama.TopicDetail{}, nil
	}
	return f.topics, nil
}

func (f *fakeAdmin) Close() error {
	f.closed = true
	return f.closeErr
}

func TestRegisterLifecycle_StartStop(t *testing.T) {
	producer := &fakeProducer{}
	admin := &fakeAdmin{}
	cfg := &Config{Brokers: []string{"localhost:9092"}, ClientID: "podzone-auth"}

	app := fxtest.New(
		t,
		fx.Supply(cfg),
		fx.Provide(
			func() Producer { return producer },
			func() Admin { return admin },
			func() pdlog.Logger { return pdlog.NopLogger{} },
		),
		fx.Invoke(registerLifecycle),
	)

	app.RequireStart()
	app.RequireStop()

	assert.True(t, producer.closed)
	assert.True(t, admin.closed)
}

func TestRegisterLifecycle_StartFailure(t *testing.T) {
	producer := &fakeProducer{}
	admin := &fakeAdmin{describeErr: errors.New("boom")}
	cfg := &Config{Brokers: []string{"localhost:9092"}, ClientID: "podzone-auth"}

	app := fx.New(
		fx.Supply(cfg),
		fx.Provide(
			func() Producer { return producer },
			func() Admin { return admin },
			func() pdlog.Logger { return pdlog.NopLogger{} },
		),
		fx.Invoke(registerLifecycle),
	)

	err := app.Start(context.Background())
	require.Error(t, err)
}
