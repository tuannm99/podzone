package common

import (
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

func AssertEqual(expected, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		log.Fatalf("assertEqual failed: expected %#v, got %#v", expected, actual)
	}
}
