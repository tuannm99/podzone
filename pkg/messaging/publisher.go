package messaging

import "context"

type Publisher interface {
	Publish(ctx context.Context, topic string, key string, msg Envelope) error
	PublishBatch(ctx context.Context, topic string, msgs []PublishRequest) error
}

type PublishRequest struct {
	Topic string
	Key   string
	Msg   Envelope
}
