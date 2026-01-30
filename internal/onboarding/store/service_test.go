package store

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/tuannm99/podzone/pkg/testkit"
)

func setupStoreService(t *testing.T) (*StoreService, *mongo.Collection, context.Context) {
	t.Helper()
	client := testkit.MongoClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)

	collection := client.Database("podzone").Collection("stores")
	_, err := collection.DeleteMany(ctx, bson.M{})
	require.NoError(t, err)

	svc := NewStoreService(StoreServiceParams{MongoClient: client})
	return svc, collection, ctx
}

func TestCreateStore_ReturnsErrSubdomainTaken_WhenExistingStoreFound(t *testing.T) {
	svc, col, ctx := setupStoreService(t)

	existing := Store{
		ID:        primitive.NewObjectID(),
		Name:      "Existing",
		Subdomain: "taken",
		OwnerID:   "owner-1",
		Status:    StoreStatusActive,
	}

	_, err := col.InsertOne(ctx, existing)
	require.NoError(t, err)

	_, err = svc.CreateStore(ctx, "New", "taken", "owner-2")
	require.ErrorIs(t, err, ErrSubdomainTaken)
}

func TestCreateStore_Success(t *testing.T) {
	svc, col, ctx := setupStoreService(t)

	store, err := svc.CreateStore(ctx, "New", "new", "owner-1")
	require.NoError(t, err)
	require.NotNil(t, store)
	require.NotZero(t, store.ID)

	count, err := col.CountDocuments(ctx, bson.M{"subdomain": "new"})
	require.NoError(t, err)
	require.Equal(t, int64(1), count)
}
