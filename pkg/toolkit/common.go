package toolkit

import (
	"encoding/json"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
)

func GetEnv(key string, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}

func GetEnvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return fallback
}

func GetEnvBool(key string, fallback bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return fallback
}

func GetEnvSlice(key string, fallback []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return fallback
}

func IsDevelopment() bool {
	return GetEnv("APP_ENV", "development") == "development"
}

func IsProduction() bool {
	return GetEnv("APP_ENV", "development") == "production"
}

func IsStaging() bool {
	return GetEnv("APP_ENV", "development") == "staging"
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
