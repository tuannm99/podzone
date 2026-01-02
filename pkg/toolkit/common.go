package toolkit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func GetEnv[T any](envName string, fallback T) T {
	v := os.Getenv(envName)
	if v == "" {
		return fallback
	}

	switch any(fallback).(type) {
	case string:
		return any(v).(T)
	case int:
		if i, err := strconv.Atoi(v); err == nil {
			return any(i).(T)
		}
	case bool:
		if b, err := strconv.ParseBool(v); err == nil {
			return any(b).(T)
		}
	}
	return fallback
}

func MapStruct[S any, T any](source S) (*T, error) {
	var target T

	data, err := json.Marshal(source)
	if err != nil {
		return nil, fmt.Errorf("err marshal from source %w", err)
	}

	err = json.Unmarshal(data, &target)
	if err != nil {
		return nil, fmt.Errorf("err unmarshal to target %w", err)
	}

	return &target, nil
}

// test helper
func MakeConfigTestDir(t *testing.T, config string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")
	require.NoError(t, os.WriteFile(path, []byte(config), 0o644))
	t.Setenv("CONFIG_PATH", path)
	return path
}

func ParseInt(s string, def int) int {
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}
