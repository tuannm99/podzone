package pdconfig

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_configProvider(t *testing.T) {
	// Always isolate env provider to avoid reading unrelated env vars from the test process.
	t.Setenv("ENV_PREFIX", "TEST")

	// 1) ENV-only mode (no CONFIG_PATH)
	k, err := NewAppConfig()
	require.NoError(t, err)
	assert.NotNil(t, k)

	// 2) YAML mode (CONFIG_PATH points to an existing file)
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yml")
	yaml := []byte(`
postgres:
  auth:
    uri: postgres://file-user:file-pass@localhost:5432/filedb
`)
	require.NoError(t, os.WriteFile(cfgPath, yaml, 0o600))

	t.Setenv("CONFIG_PATH", cfgPath)

	k, err = NewAppConfig()
	require.NoError(t, err)
	assert.NotNil(t, k)

	// Ensure file value is read
	assert.Equal(t,
		"postgres://file-user:file-pass@localhost:5432/filedb",
		k.String("postgres.auth.uri"),
	)
}

func TestModule_EnvOverridesFile(t *testing.T) {
	// Isolate env provider to only read envs with TEST_ prefix.
	t.Setenv("ENV_PREFIX", "TEST")

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yml")
	yaml := []byte(`
postgres:
  auth:
    uri: postgres://file-user:file-pass@localhost:5432/filedb
`)
	require.NoError(t, os.WriteFile(cfgPath, yaml, 0o600))

	t.Setenv("CONFIG_PATH", cfgPath)

	// Env override should win because NewAppConfig() loads env after YAML.
	t.Setenv("TEST_POSTGRES_AUTH_URI", "postgres://env-user:env-pass@localhost:5432/envdb")

	k, err := NewAppConfig()
	require.NoError(t, err)
	assert.NotNil(t, k)

	assert.Equal(t,
		"postgres://env-user:env-pass@localhost:5432/envdb",
		k.String("postgres.auth.uri"),
	)
}
