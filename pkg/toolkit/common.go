package toolkit

import (
	"encoding/json"
	"log"
	"os"
	"reflect"
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

func AssertEqual(expected, actual any) {
	if !reflect.DeepEqual(expected, actual) {
		log.Fatalf("assertEqual failed: expected %#v, got %#v", expected, actual)
	}
}

func MapStruct[S any, T any](source S) *T {
	var target T

	data, err := json.Marshal(source)
	if err != nil {
		return nil
	}

	err = json.Unmarshal(data, &target)
	if err != nil {
		return nil
	}

	return &target
}
