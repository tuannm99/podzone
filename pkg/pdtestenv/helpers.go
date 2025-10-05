package pdtestenv

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// MakeConfigDir writes a temporary config file and sets CONFIG_PATH
func MakeConfigDir(t *testing.T, config string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")
	require.NoError(t, os.WriteFile(path, []byte(config), 0o644))
	t.Setenv("CONFIG_PATH", path)
	return path
}
