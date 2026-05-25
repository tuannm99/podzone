package messaging

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeSagaStep struct {
	name        string
	executeErr  error
	executed    *[]string
	compensated *[]string
}

func (f *fakeSagaStep) Name() string { return f.name }

func (f *fakeSagaStep) Execute(ctx context.Context, msg Envelope) error {
	*f.executed = append(*f.executed, f.name)
	return f.executeErr
}

func (f *fakeSagaStep) Compensate(ctx context.Context, msg Envelope) error {
	*f.compensated = append(*f.compensated, f.name)
	return nil
}

func TestSagaCompensatesCompletedSteps(t *testing.T) {
	executed := make([]string, 0, 3)
	compensated := make([]string, 0, 3)
	saga := NewSaga(
		&fakeSagaStep{name: "one", executed: &executed, compensated: &compensated},
		&fakeSagaStep{name: "two", executed: &executed, compensated: &compensated},
		&fakeSagaStep{name: "three", executeErr: errors.New("boom"), executed: &executed, compensated: &compensated},
	)

	err := saga.Run(context.Background(), Envelope{ID: "evt_1"})
	require.Error(t, err)
	assert.Equal(t, []string{"one", "two", "three"}, executed)
	assert.Equal(t, []string{"two", "one"}, compensated)
}
