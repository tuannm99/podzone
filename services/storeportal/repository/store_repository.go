package repository

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tuannm99/podzone/services/storeportal/models"
)

var ErrStoreNotFound = errors.New("store not found")

// StoreRepository handles database operations for stores
type StoreRepository struct {
	collection *mongo.Collection
}

// NewStoreRepository creates a new store repository
func NewStoreRepository(db *mongo.Client) *StoreRepository {
	// TODO : switch database base on context -> provide for each repository
	// Make a connection pooling for switching context each tenant Pool 1000
	collection := db.Database("storeportal").Collection("stores")
	return &StoreRepository{
		collection: collection,
	}
}

// Create creates a new store
func (r *StoreRepository) Create(ctx context.Context, store *models.Store) error {
	result, err := r.collection.InsertOne(ctx, store)
	if err != nil {
		return err
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		store.ID = oid
	}
	return nil
}

// Get retrieves a store by ID
func (r *StoreRepository) Get(ctx context.Context, id string) (*models.Store, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var store models.Store
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&store)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrStoreNotFound
		}
		return nil, err
	}

	return &store, nil
}

// ListByOwnerID retrieves all stores for an owner
func (r *StoreRepository) ListByOwnerID(ctx context.Context, ownerID string) ([]*models.Store, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"owner_id": ownerID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var stores []*models.Store
	if err := cursor.All(ctx, &stores); err != nil {
		return nil, err
	}

	return stores, nil
}

// Update updates a store
func (r *StoreRepository) Update(ctx context.Context, store *models.Store) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": store.ID},
		bson.M{"$set": store},
		options.Update().SetUpsert(false),
	)
	return err
}
