package pdmongo

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

// mongo.<name>.*
type Config struct {
	Provider string `mapstructure:"provider"` // "real" | "mock"
	URI      string `mapstructure:"uri"`
}

type ProviderFn func(ctx context.Context, cfg Config) (*mongo.Client, error)
