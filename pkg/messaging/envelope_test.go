package messaging

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvelopeValidate(t *testing.T) {
	validPayload := json.RawMessage(`{"ok":true}`)
	valid := Envelope{
		ID:            "evt_1",
		Type:          "tenant.created",
		Source:        "iam",
		OccurredAt:    time.Now().UTC(),
		SchemaVersion: 1,
		Payload:       validPayload,
	}
	require.NoError(t, valid.Validate())

	tests := []struct {
		name string
		env  Envelope
		err  error
	}{
		{"missing id", Envelope{Type: "x", Source: "y", Payload: validPayload}, ErrEmptyEnvelopeID},
		{"missing type", Envelope{ID: "1", Source: "y", Payload: validPayload}, ErrEmptyEnvelopeType},
		{"missing source", Envelope{ID: "1", Type: "x", Payload: validPayload}, ErrEmptyEnvelopeSource},
		{"missing payload", Envelope{ID: "1", Type: "x", Source: "y"}, ErrEmptyEnvelopePayload},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.ErrorIs(t, tt.env.Validate(), tt.err)
		})
	}
}
