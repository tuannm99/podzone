package pdmongo

import (
	"context"
	"time"

	"github.com/tuannm99/podzone/pkg/toolkit"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type InstanceConfig struct {
	URI string `mapstructure:"uri"`
}

type MongoFactory = toolkit.Factory[*mongo.Client, InstanceConfig]

var Registry = toolkit.NewRegistry[*mongo.Client, InstanceConfig]("real")

var currentFactoryID = "real"

func init() {
	Registry.Register("real", RealMongoFactory)
	Registry.Register("noop", NoopMongoFactory)
}

// ---- Factories ----

var RealMongoFactory MongoFactory = func(ctx context.Context, cfg InstanceConfig) (*mongo.Client, error) {
	cl, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.URI))
	if err != nil {
		return nil, err
	}
	return cl, nil
}

var NoopMongoFactory MongoFactory = func(ctx context.Context, conf InstanceConfig) (*mongo.Client, error) {
	return mongo.Connect(ctx, options.Client().ApplyURI(conf.URI))
}

func ping(ctx context.Context, c *mongo.Client) error {
	tctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return c.Ping(tctx, nil)
}
