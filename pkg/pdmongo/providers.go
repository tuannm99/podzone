package pdmongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var RealProvider ProviderFn = func(ctx context.Context, cfg Config) (*mongo.Client, error) {
	cl, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.URI))
	if err != nil {
		return nil, err
	}
	return cl, nil
}

var MockProvider ProviderFn = func(ctx context.Context, cfg Config) (*mongo.Client, error) {
	uri := cfg.URI
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}
	cl, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	_ = ctx
	_ = time.Second
	return cl, nil
}
