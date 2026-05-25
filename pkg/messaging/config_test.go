package messaging

import (
	"testing"
	"time"

	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
)

func TestLoadConsumerRuntimeConfig_Defaults(t *testing.T) {
	cfg := LoadConsumerRuntimeConfig(nil, "", DefaultConsumerRuntimeConfig("auth.iam-projection"))

	assert.True(t, cfg.Enabled)
	assert.Equal(t, 5, cfg.MaxAttempts)
	assert.Equal(t, time.Second, cfg.BaseDelay)
	assert.Equal(t, "auth.iam-projection", cfg.Idempotency.ConsumerName)
	assert.Equal(t, "message_inbox", cfg.Idempotency.TableName)
}

func TestLoadConsumerRuntimeConfig_FromKoanf(t *testing.T) {
	k := koanf.New(".")
	k.Set("messaging.auth.consumers.iam_projection.max_attempts", 7)
	k.Set("messaging.auth.consumers.iam_projection.base_delay", "3s")
	k.Set("messaging.auth.consumers.iam_projection.observability.enabled", true)
	k.Set("messaging.auth.consumers.iam_projection.idempotency.enabled", true)
	k.Set("messaging.auth.consumers.iam_projection.idempotency.table_name", "auth_inbox")

	cfg := LoadConsumerRuntimeConfig(
		k,
		"messaging.auth.consumers.iam_projection",
		DefaultConsumerRuntimeConfig("auth.iam-projection"),
	)

	assert.Equal(t, 7, cfg.MaxAttempts)
	assert.Equal(t, 3*time.Second, cfg.BaseDelay)
	assert.True(t, cfg.Observability.Enabled)
	assert.True(t, cfg.Idempotency.Enabled)
	assert.Equal(t, "auth_inbox", cfg.Idempotency.TableName)
}
