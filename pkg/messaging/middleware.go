package messaging

import "context"

type HandlerFunc func(ctx context.Context, msg Envelope) error

func (f HandlerFunc) Handle(ctx context.Context, msg Envelope) error {
	return f(ctx, msg)
}

type Middleware func(next Handler) Handler

func Chain(handler Handler, middlewares ...Middleware) Handler {
	chained := handler
	for i := len(middlewares) - 1; i >= 0; i-- {
		if middlewares[i] == nil {
			continue
		}
		chained = middlewares[i](chained)
	}
	return chained
}
