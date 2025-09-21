package pdconfig

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_configProvider(t *testing.T) {
	v, err := configProvider()

	require.NoError(t, err)
	assert.NotNil(t, v)

	t.Setenv("CONFIG_PATH", "config.yml")
	v, err = configProvider()

	require.NoError(t, err)
	assert.NotNil(t, v)
}

func TestModule_EnvOverridesFile(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yml")
	yaml := []byte(`
postgres:
  auth:
    uri: postgres://file-user:file-pass@localhost:5432/filedb
`)
	if err := os.WriteFile(cfgPath, yaml, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	t.Setenv("CONFIG_PATH", cfgPath)
	t.Setenv("POSTGRES_AUTH_URI", "postgres://env-user:env-pass@localhost:5432/envdb")

	v, err := configProvider()
	require.NoError(t, err)
	assert.NotNil(t, v)
}
