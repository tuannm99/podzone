package messaging

import "context"

type Handler interface {
	Handle(ctx context.Context, msg Envelope) error
}
