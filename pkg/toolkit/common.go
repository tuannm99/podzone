package toolkit

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
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
