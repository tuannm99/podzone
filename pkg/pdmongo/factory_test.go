package pdmongo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewMongoDbFromConfig(t *testing.T) {
	cfg := &Config{
		URI:            "mongodb://localhost:27017",
		ConnectTimeout: time.Millisecond * 1,
		PingTimeout:    time.Millisecond * 1,
	}

	_, err := NewMongoDbFromConfig(cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "pdmongo: ping failed")
}
