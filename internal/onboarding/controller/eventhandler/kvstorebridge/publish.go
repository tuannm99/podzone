package kvstorebridge

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tuannm99/podzone/internal/onboarding/infrastructure/messaging/publisher"
	"github.com/tuannm99/podzone/pkg/messaging"
)

type PublishHandler struct {
	pub *publisher.KVStorePublisher
}

var _ messaging.TypedHandler = (*PublishHandler)(nil)

func NewPublishHandler(pub *publisher.KVStorePublisher) *PublishHandler {
	return &PublishHandler{pub: pub}
}

func (h *PublishHandler) MessageType() string {
	return "kv_store.publish"
}

func (h *PublishHandler) Handle(ctx context.Context, msg messaging.Envelope) error {
	var payload struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return messaging.DeadLetterError(
			fmt.Errorf("decode kv_store.publish payload: %w", err),
			"invalid kv_store.publish payload",
		)
	}
	if payload.Key == "" {
		return messaging.DeadLetterError(
			fmt.Errorf("kv_store.publish payload missing key"),
			"missing kv store key",
		)
	}
	if err := h.pub.Put(ctx, payload.Key, payload.Value); err != nil {
		return messaging.RetryableError(
			fmt.Errorf("publish kv store key %q: %w", payload.Key, err),
			"kv store unavailable",
		)
	}
	return nil
}
