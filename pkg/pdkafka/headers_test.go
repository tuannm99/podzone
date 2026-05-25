package pdkafka

import (
	"testing"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToRecordHeadersReturnsNilForEmptyInput(t *testing.T) {
	assert.Nil(t, ToRecordHeaders(nil))
	assert.Nil(t, ToRecordHeaders(map[string]string{}))
}

func TestToRecordHeadersPreservesEntries(t *testing.T) {
	headers := ToRecordHeaders(map[string]string{
		"tenant_id": "t1",
		"type":      "tenant.created",
	})

	require.Len(t, headers, 2)
	values := map[string]string{}
	for _, item := range headers {
		values[string(item.Key)] = string(item.Value)
	}
	assert.Equal(t, "t1", values["tenant_id"])
	assert.Equal(t, "tenant.created", values["type"])
}

func TestFromRecordHeadersPreservesEntries(t *testing.T) {
	headers := FromRecordHeaders([]*sarama.RecordHeader{
		{Key: []byte("tenant_id"), Value: []byte("t1")},
		{Key: []byte("type"), Value: []byte("tenant.created")},
	})

	require.Len(t, headers, 2)
	assert.Equal(t, "t1", headers["tenant_id"])
	assert.Equal(t, "tenant.created", headers["type"])
}
