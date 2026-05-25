package messaging

import "context"

type DeliveryEvent struct {
	ConsumerName string
	Topic        string
	Key          string
	Envelope     Envelope
	Action       FailureAction
	Reason       string
	Err          error
}

type ObserverFunc func(ctx context.Context, event DeliveryEvent)

func (f ObserverFunc) Observe(ctx context.Context, event DeliveryEvent) {
	f(ctx, event)
}
