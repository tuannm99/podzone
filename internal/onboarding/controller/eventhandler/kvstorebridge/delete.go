package kvstorebridge

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tuannm99/podzone/internal/onboarding/infrastructure/messaging/publisher"
	"github.com/tuannm99/podzone/pkg/messaging"
)

type DeleteHandler struct {
	pub *publisher.KVStorePublisher
}

var _ messaging.TypedHandler = (*DeleteHandler)(nil)

func NewDeleteHandler(pub *publisher.KVStorePublisher) *DeleteHandler {
	return &DeleteHandler{pub: pub}
}

func (h *DeleteHandler) MessageType() string {
	return "kv_store.delete"
}

func (h *DeleteHandler) Handle(ctx context.Context, msg messaging.Envelope) error {
	var payload struct {
		Key string `json:"key"`
	}
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return messaging.DeadLetterError(
			fmt.Errorf("decode kv_store.delete payload: %w", err),
			"invalid kv_store.delete payload",
		)
	}
	if payload.Key == "" {
		return messaging.DeadLetterError(
			fmt.Errorf("kv_store.delete payload missing key"),
			"missing kv store key",
		)
	}
	if err := h.pub.Delete(ctx, payload.Key); err != nil {
		return messaging.RetryableError(
			fmt.Errorf("delete kv store key %q: %w", payload.Key, err),
			"kv store unavailable",
		)
	}
	return nil
}
