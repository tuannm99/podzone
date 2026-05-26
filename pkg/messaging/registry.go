package messaging

import (
	"context"
	"fmt"
)

type Registry struct {
	handlers map[string]Handler
}

var _ Handler = (*Registry)(nil)

func NewRegistry(handlers ...TypedHandler) (*Registry, error) {
	registry := &Registry{handlers: make(map[string]Handler, len(handlers))}
	for _, handler := range handlers {
		if handler == nil {
			continue
		}
		messageType := handler.MessageType()
		if messageType == "" {
			return nil, fmt.Errorf("messaging: empty message type")
		}
		if _, exists := registry.handlers[messageType]; exists {
			return nil, fmt.Errorf("messaging: duplicate handler for %q", messageType)
		}
		registry.handlers[messageType] = handler
	}
	return registry, nil
}

func (r *Registry) Handle(ctx context.Context, msg Envelope) error {
	if r == nil {
		return nil
	}
	handler, ok := r.handlers[msg.Type]
	if !ok {
		return nil
	}
	return handler.Handle(ctx, msg)
}
