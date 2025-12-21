package pdmongo

import (
	"testing"

	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/require"
)

func TestGetConfigFromKoanf_NoSub(t *testing.T) {
	k := koanf.New(".")

	cfg, err := GetConfig("notfound", k)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// empty because mongo.notfound.* does not exist
	require.Empty(t, cfg.URI)
	require.Empty(t, cfg.Database)
}

func TestGetConfigFromKoanf_WithSub(t *testing.T) {
	k := koanf.New(".")

	// Koanf stores values by key path
	k.Set("mongo.test.uri", "mongodb://localhost:27017")
	k.Set("mongo.test.database", "catalog")
	k.Set("mongo.test.connect_timeout", "5s")
	k.Set("mongo.test.ping_timeout", "10s")

	cfg, err := GetConfig("test", k)
	require.NoError(t, err)
	require.Equal(t, "mongodb://localhost:27017", cfg.URI)
	require.Equal(t, "catalog", cfg.Database)
}
