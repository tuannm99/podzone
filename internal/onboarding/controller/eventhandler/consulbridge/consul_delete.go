package consulbridge

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tuannm99/podzone/internal/onboarding/infrastructure/messaging/publisher"
	"github.com/tuannm99/podzone/pkg/messaging"
)

type ConsulDeleteHandler struct {
	pub *publisher.ConsulPublisher
}

var _ messaging.TypedHandler = (*ConsulDeleteHandler)(nil)

func NewConsulDeleteHandler(pub *publisher.ConsulPublisher) *ConsulDeleteHandler {
	return &ConsulDeleteHandler{pub: pub}
}

func (h *ConsulDeleteHandler) MessageType() string {
	return "consul.delete"
}

func (h *ConsulDeleteHandler) Handle(ctx context.Context, msg messaging.Envelope) error {
	var payload struct {
		Key string `json:"key"`
	}
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return messaging.DeadLetterError(
			fmt.Errorf("decode consul.delete payload: %w", err),
			"invalid consul.delete payload",
		)
	}
	if payload.Key == "" {
		return messaging.DeadLetterError(
			fmt.Errorf("consul.delete payload missing key"),
			"missing consul key",
		)
	}
	if err := h.pub.Delete(ctx, payload.Key); err != nil {
		return messaging.RetryableError(
			fmt.Errorf("delete consul key %q: %w", payload.Key, err),
			"consul unavailable",
		)
	}
	return nil
}
