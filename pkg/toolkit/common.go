package toolkit

import (
	"encoding/json"
	"log"
	"os"
	"reflect"
)

func FallbackEnv(key string, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
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
