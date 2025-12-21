package pdredis

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetConfigFromViper(t *testing.T) {
	v := viper.New()
	v.Set("redis.test.uri", "redis://localhost:6379/0")

	cfg, err := GetConfigFromViper("test", v)
	require.NoError(t, err)
	assert.Equal(t, "redis://localhost:6379/0", cfg.URI)
}

func TestNewClientFromConfig_ValidURI(t *testing.T) {
	cfg := &Config{URI: "redis://:pass@localhost:6379/2"}
	client, err := NewClientFromConfig(cfg)
	require.NoError(t, err)
	require.NotNil(t, client)

	opts := client.Options()
	assert.Equal(t, "localhost:6379", opts.Addr)
	assert.Equal(t, "pass", opts.Password)
	assert.Equal(t, 2, opts.DB)
}

func TestNewClientFromConfig_NilConfig(t *testing.T) {
	client, err := NewClientFromConfig(nil)
	require.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "nil redis config")
}

func TestNewClientFromConfig_InvalidURI(t *testing.T) {
	t.Parallel()

	cfg := &Config{URI: "redis://localhost:abc/0"}
	client, err := NewClientFromConfig(cfg)
	require.Error(t, err)
	require.Nil(t, client)
}
