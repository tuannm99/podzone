package kvstores

import (
	"errors"
	"fmt"

	"github.com/knadh/koanf/v2"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/toolkit/kvstores"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/fx"
)

type Config struct {
	Provider   string `mapstructure:"provider"`
	Database   string `mapstructure:"database"`
	Collection string `mapstructure:"collection"`
}

func NewConfigFromKoanf(k *koanf.Koanf) (Config, error) {
	cfg := Config{
		Provider:   "mongo",
		Database:   "onboarding",
		Collection: "runtime_kv",
	}
	if k == nil {
		return cfg, errors.New("koanf is nil")
	}
	if err := k.Unmarshal("kv_store", &cfg); err != nil {
		return cfg, fmt.Errorf("unmarshal kv_store config: %w", err)
	}
	if cfg.Provider != "mongo" {
		return cfg, fmt.Errorf("unsupported kv_store provider %q", cfg.Provider)
	}
	if cfg.Database == "" {
		return cfg, errors.New("missing config: kv_store.database")
	}
	if cfg.Collection == "" {
		return cfg, errors.New("missing config: kv_store.collection")
	}
	return cfg, nil
}

func NewMongoStoreFromConfig(
	logger pdlog.Logger,
	client *mongo.Client,
	cfg Config,
) (*kvstores.MongoStore, error) {
	return kvstores.NewMongoStore(client, cfg.Database, cfg.Collection, logger)
}

func ModuleFor(mongoName string) fx.Option {
	clientTag := `name:"mongo-` + mongoName + `"`
	return fx.Provide(
		NewConfigFromKoanf,
		fx.Annotate(
			NewMongoStoreFromConfig,
			fx.ParamTags(``, clientTag, ``),
			fx.As(new(kvstores.KVStore)),
		),
	)
}
