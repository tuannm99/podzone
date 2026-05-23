package messaging

import (
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type Envelope struct {
	ID            string            `json:"id"`
	Type          string            `json:"type"`
	Source        string            `json:"source"`
	TenantID      string            `json:"tenant_id,omitempty"`
	EntityID      string            `json:"entity_id,omitempty"`
	CorrelationID string            `json:"correlation_id,omitempty"`
	CausationID   string            `json:"causation_id,omitempty"`
	OccurredAt    time.Time         `json:"occurred_at"`
	SchemaVersion int               `json:"schema_version"`
	Headers       map[string]string `json:"headers,omitempty"`
	Payload       json.RawMessage   `json:"payload"`
}

var (
	ErrEmptyEnvelopeID      = errors.New("messaging: envelope id is required")
	ErrEmptyEnvelopeType    = errors.New("messaging: envelope type is required")
	ErrEmptyEnvelopeSource  = errors.New("messaging: envelope source is required")
	ErrEmptyEnvelopePayload = errors.New("messaging: envelope payload is required")
)

func (e Envelope) Validate() error {
	switch {
	case strings.TrimSpace(e.ID) == "":
		return ErrEmptyEnvelopeID
	case strings.TrimSpace(e.Type) == "":
		return ErrEmptyEnvelopeType
	case strings.TrimSpace(e.Source) == "":
		return ErrEmptyEnvelopeSource
	case len(e.Payload) == 0:
		return ErrEmptyEnvelopePayload
	default:
		return nil
	}
}
