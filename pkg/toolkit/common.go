package toolkit

import (
	"log"
	"os"
	"reflect"

	"github.com/mitchellh/mapstructure"
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
	_ = mapstructure.Decode(source, &target)
	return &target
}
