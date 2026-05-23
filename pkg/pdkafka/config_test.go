package pdkafka

import (
	"testing"

	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetConfigFromKoanf_Defaults(t *testing.T) {
	k := koanf.New(".")
	k.Set("kafka.auth.brokers", []string{"localhost:9092"})

	cfg, err := GetConfigFromKoanf("auth", k)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, []string{"localhost:9092"}, cfg.Brokers)
	assert.Equal(t, "podzone-auth", cfg.ClientID)
	assert.Equal(t, "podzone.auth", cfg.ConsumerGroupPrefix)
	assert.Equal(t, "3.7.0", cfg.Version)
	assert.Equal(t, RequiredAcksAll, cfg.RequiredAcks)
	assert.Equal(t, CompressionZSTD, cfg.Compression)
}

func TestGetConfigFromKoanf_RequiresBrokers(t *testing.T) {
	k := koanf.New(".")

	cfg, err := GetConfigFromKoanf("auth", k)
	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "brokers is required")
}
