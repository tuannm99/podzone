package infrastructure

import (
	"go.uber.org/fx"

	"github.com/tuannm99/podzone/services/storeportal/domain/repositories"
	"github.com/tuannm99/podzone/services/storeportal/infrastructure/persistence/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
)

// Module exports all infrastructure layer components
var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			provideMongoStoreRepository,
			fx.ResultTags(`name:"store-repository"`),
		),
	),
)

// MongoParams contains MongoDB connection parameters
type MongoParams struct {
	fx.In

	MongoClient *mongo.Client `name:"mongo-storeportal"`
}

// provideMongoStoreRepository creates a new MongoDB store repository
func provideMongoStoreRepository(params MongoParams) repositories.StoreRepository {
	return mongodb.NewStoreRepository(params.MongoClient, "podzone", "stores")
}
