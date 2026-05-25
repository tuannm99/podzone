package messaging

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChainMiddleware(t *testing.T) {
	calls := make([]string, 0, 4)
	handler := HandlerFunc(func(ctx context.Context, msg Envelope) error {
		calls = append(calls, "handle")
		return nil
	})
	m1 := func(next Handler) Handler {
		return HandlerFunc(func(ctx context.Context, msg Envelope) error {
			calls = append(calls, "m1.before")
			err := next.Handle(ctx, msg)
			calls = append(calls, "m1.after")
			return err
		})
	}
	m2 := func(next Handler) Handler {
		return HandlerFunc(func(ctx context.Context, msg Envelope) error {
			calls = append(calls, "m2.before")
			err := next.Handle(ctx, msg)
			calls = append(calls, "m2.after")
			return err
		})
	}

	err := Chain(handler, m1, m2).Handle(context.Background(), Envelope{ID: "evt_1"})
	assert.NoError(t, err)
	assert.Equal(t, []string{"m1.before", "m2.before", "handle", "m2.after", "m1.after"}, calls)
}
