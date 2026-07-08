package sqlstore

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tuannm99/podzone/pkg/messaging"
)

func TestNewOutboxStore_DefaultTableName(t *testing.T) {
	store, err := NewOutboxStore(nil, "")
	require.NoError(t, err)
	require.NotNil(t, store)
	assert.Equal(t, "message_outbox", store.tableName)
}

func TestNewOutboxStore_RejectsInvalidIdentifier(t *testing.T) {
	store, err := NewOutboxStore(nil, "message_outbox;drop table users")
	require.Error(t, err)
	assert.Nil(t, store)
}

func TestNewOutboxStore_AllowsOwnerTableName(t *testing.T) {
	store, err := NewOutboxStore(nil, "iam_outbox")
	require.NoError(t, err)
	require.NotNil(t, store)
	assert.Equal(t, "iam_outbox", store.tableName)
}

func TestOutboxRowToRecord(t *testing.T) {
	now := time.Date(2026, 7, 9, 10, 0, 0, 0, time.UTC)
	envelope := messaging.Envelope{
		ID:            "evt_1",
		Type:          "iam.user.created",
		Source:        "iam",
		OccurredAt:    now,
		SchemaVersion: 1,
		Payload:       json.RawMessage(`{"user_id":"u1"}`),
	}
	payload, err := json.Marshal(envelope)
	require.NoError(t, err)

	record, err := outboxRow{
		ID:            "evt_1",
		Topic:         "podzone.iam.events",
		MessageKey:    "u1",
		EnvelopeJSON:  payload,
		Status:        "pending",
		Attempts:      2,
		NextAttemptAt: now,
		CreatedAt:     now,
		UpdatedAt:     now,
	}.toRecord()

	require.NoError(t, err)
	assert.Equal(t, "evt_1", record.ID)
	assert.Equal(t, "podzone.iam.events", record.Topic)
	assert.Equal(t, envelope, record.Envelope)
	assert.Equal(t, 2, record.Attempts)
}
