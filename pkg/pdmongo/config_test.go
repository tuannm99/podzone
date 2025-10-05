package pdmongo

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestGetConfigFromViper_NoSub(t *testing.T) {
	v := viper.New()
	cfg, err := GetConfigFromViper("notfound", v)
	require.NoError(t, err)
	require.NotNil(t, cfg)
}

func TestGetConfigFromViper_WithSub(t *testing.T) {
	v := viper.New()
	v.Set("mongo.test.uri", "mongodb://localhost:27017")
	v.Set("mongo.test.database", "catalog")
	v.Set("mongo.test.connect_timeout", 5)
	v.Set("mongo.test.ping_timeout", 10)

	cfg, err := GetConfigFromViper("test", v)
	require.NoError(t, err)
	require.Equal(t, "mongodb://localhost:27017", cfg.URI)
	require.Equal(t, "catalog", cfg.Database)
}
