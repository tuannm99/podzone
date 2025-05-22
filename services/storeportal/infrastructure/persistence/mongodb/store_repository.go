package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tuannm99/podzone/services/storeportal/domain/entities"
)

// StoreRepository implements the domain StoreRepository interface using MongoDB
type StoreRepository struct {
	collection *mongo.Collection
}

// NewStoreRepository creates a new MongoDB store repository
func NewStoreRepository(client *mongo.Client, dbName, collectionName string) *StoreRepository {
	collection := client.Database(dbName).Collection(collectionName)

	// Create indexes
	ctx := context.Background()
	_, _ = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "name", Value: 1}},
		Options: options.Index().SetUnique(true),
	})

	return &StoreRepository{
		collection: collection,
	}
}

// Create creates a new store
func (r *StoreRepository) Create(ctx context.Context, store *entities.Store) error {
	store.ID = primitive.NewObjectID().Hex()
	_, err := r.collection.InsertOne(ctx, store)
	return err
}

// Get retrieves a store by ID
func (r *StoreRepository) Get(ctx context.Context, id string) (*entities.Store, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var store entities.Store
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&store)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &store, nil
}

// List retrieves all stores
func (r *StoreRepository) List(ctx context.Context) ([]*entities.Store, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var stores []*entities.Store
	if err = cursor.All(ctx, &stores); err != nil {
		return nil, err
	}

	return stores, nil
}

// ListByOwnerID retrieves all stores for a specific owner
func (r *StoreRepository) ListByOwnerID(ctx context.Context, ownerID string) ([]*entities.Store, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"owner_id": ownerID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var stores []*entities.Store
	if err = cursor.All(ctx, &stores); err != nil {
		return nil, err
	}

	return stores, nil
}

// Update updates an existing store
func (r *StoreRepository) Update(ctx context.Context, store *entities.Store) error {
	objectID, err := primitive.ObjectIDFromHex(store.ID)
	if err != nil {
		return err
	}

	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": store},
	)
	return err
}

// Delete deletes a store by ID
func (r *StoreRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	return err
}
