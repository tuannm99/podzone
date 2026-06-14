package repository

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/entity"
)

func TestOutboxMessageToRecord_SeparatesKafkaTopicFromMessageType(t *testing.T) {
	record, err := outboxMessageToRecord(entity.OutboxMessage{
		EventID:   "event-1",
		Topic:     "consul.publish",
		TenantID:  "tenant-1",
		InfraType: entity.InfraPostgres,
		Name:      "primary",
		Payload: map[string]interface{}{
			"key":   "podzone/tenants/tenant-1/placement",
			"value": "{}",
		},
	})

	require.NoError(t, err)
	require.Equal(t, "podzone.onboarding.events", record.Topic)
	require.Equal(t, "consul.publish", record.Envelope.Type)
	require.Equal(t, "tenant-1:postgres:primary", record.MessageKey)
}
