package consulbridge

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tuannm99/podzone/internal/onboarding/infrasmanager/infrastructure/publisher"
	"github.com/tuannm99/podzone/pkg/messaging"
)

type ConsulPublishHandler struct {
	pub *publisher.ConsulPublisher
}

var _ messaging.TypedHandler = (*ConsulPublishHandler)(nil)

func NewConsulPublishHandler(pub *publisher.ConsulPublisher) *ConsulPublishHandler {
	return &ConsulPublishHandler{pub: pub}
}

func (h *ConsulPublishHandler) MessageType() string {
	return "consul.publish"
}

func (h *ConsulPublishHandler) Handle(ctx context.Context, msg messaging.Envelope) error {
	var payload struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return messaging.DeadLetterError(
			fmt.Errorf("decode consul.publish payload: %w", err),
			"invalid consul.publish payload",
		)
	}
	if payload.Key == "" {
		return messaging.DeadLetterError(
			fmt.Errorf("consul.publish payload missing key"),
			"missing consul key",
		)
	}
	if err := h.pub.Put(ctx, payload.Key, payload.Value); err != nil {
		return messaging.RetryableError(
			fmt.Errorf("publish consul key %q: %w", payload.Key, err),
			"consul unavailable",
		)
	}
	return nil
}
