package messaging

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubTypedHandler struct {
	messageType string
	err         error
	called      bool
}

func (h *stubTypedHandler) MessageType() string {
	return h.messageType
}

func (h *stubTypedHandler) Handle(_ context.Context, _ Envelope) error {
	h.called = true
	return h.err
}

func TestNewRegistryRejectsDuplicateHandlers(t *testing.T) {
	_, err := NewRegistry(
		&stubTypedHandler{messageType: "tenant.created"},
		&stubTypedHandler{messageType: "tenant.created"},
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate handler")
}

func TestRegistryDispatchesRegisteredHandler(t *testing.T) {
	handler := &stubTypedHandler{messageType: "tenant.created"}
	registry, err := NewRegistry(handler)
	require.NoError(t, err)

	err = registry.Handle(context.Background(), Envelope{Type: "tenant.created"})
	require.NoError(t, err)
	assert.True(t, handler.called)
}

func TestRegistryRejectsUnknownMessages(t *testing.T) {
	registry, err := NewRegistry()
	require.NoError(t, err)

	err = registry.Handle(context.Background(), Envelope{Type: "unknown.event"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "no handler registered")
}

func TestRegistryRejectsNilReceiver(t *testing.T) {
	var registry *Registry

	err := registry.Handle(context.Background(), Envelope{Type: "tenant.created"})
	require.ErrorIs(t, err, ErrNilRegistry)
}

func TestRegistryPropagatesHandlerError(t *testing.T) {
	expectedErr := errors.New("boom")
	handler := &stubTypedHandler{messageType: "tenant.created", err: expectedErr}
	registry, err := NewRegistry(handler)
	require.NoError(t, err)

	err = registry.Handle(context.Background(), Envelope{Type: "tenant.created"})
	require.ErrorIs(t, err, expectedErr)
}
