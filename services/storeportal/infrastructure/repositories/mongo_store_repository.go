package repositories

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/tuannm99/podzone/pkg/contextfx"
	"github.com/tuannm99/podzone/services/storeportal/domain/entities"
)

var ErrStoreNotFound = errors.New("store not found")

// MongoStoreRepository implements StoreRepository using MongoDB
type MongoStoreRepository struct {
	client *mongo.Client
}

// NewMongoStoreRepository creates a new MongoDB store repository
func NewMongoStoreRepository(client *mongo.Client) *MongoStoreRepository {
	return &MongoStoreRepository{
		client: client,
	}
}

// getTenantDB returns the database for a specific tenant
func (r *MongoStoreRepository) GetTenantDB(ctx context.Context, tenantID string) (*mongo.Database, error) {
	// Get tenant_id from context
	ctxTenantID, ok := contextfx.GetTenantID(ctx)
	if !ok {
		return nil, contextfx.ErrTenantNotFound
	}

	// Verify that the tenant_id in context matches the requested tenant_id
	if ctxTenantID != tenantID {
		return nil, contextfx.ErrUnauthorized
	}

	// Use tenant_id as database name
	return r.client.Database(tenantID), nil
}

// Create creates a new store
func (r *MongoStoreRepository) Create(ctx context.Context, store *entities.Store) error {
	tenantID, ok := contextfx.GetTenantID(ctx)
	if !ok {
		return contextfx.ErrTenantNotFound
	}

	db := r.client.Database(tenantID)
	collection := db.Collection("stores")

	// Generate a new ObjectID
	store.ID = primitive.NewObjectID().Hex()

	// Insert the store
	_, err := collection.InsertOne(ctx, store)
	return err
}

// Get retrieves a store by ID
func (r *MongoStoreRepository) Get(ctx context.Context, id string) (*entities.Store, error) {
	tenantID, ok := contextfx.GetTenantID(ctx)
	if !ok {
		return nil, contextfx.ErrTenantNotFound
	}

	db := r.client.Database(tenantID)
	collection := db.Collection("stores")

	// Find the store
	var store entities.Store
	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&store)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrStoreNotFound
		}
		return nil, err
	}

	return &store, nil
}

// List retrieves all stores
func (r *MongoStoreRepository) List(ctx context.Context) ([]*entities.Store, error) {
	tenantID, ok := contextfx.GetTenantID(ctx)
	if !ok {
		return nil, contextfx.ErrTenantNotFound
	}

	db := r.client.Database(tenantID)
	collection := db.Collection("stores")

	// Find all stores
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var stores []*entities.Store
	if err := cursor.All(ctx, &stores); err != nil {
		return nil, err
	}

	return stores, nil
}

// ListByOwnerID retrieves all stores for a specific owner
func (r *MongoStoreRepository) ListByOwnerID(ctx context.Context, ownerID string) ([]*entities.Store, error) {
	tenantID, ok := contextfx.GetTenantID(ctx)
	if !ok {
		return nil, contextfx.ErrTenantNotFound
	}

	db := r.client.Database(tenantID)
	collection := db.Collection("stores")

	// Find stores by owner_id
	cursor, err := collection.Find(ctx, bson.M{"owner_id": ownerID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var stores []*entities.Store
	if err := cursor.All(ctx, &stores); err != nil {
		return nil, err
	}

	return stores, nil
}

// Update updates an existing store
func (r *MongoStoreRepository) Update(ctx context.Context, store *entities.Store) error {
	tenantID, ok := contextfx.GetTenantID(ctx)
	if !ok {
		return contextfx.ErrTenantNotFound
	}

	db := r.client.Database(tenantID)
	collection := db.Collection("stores")

	// Update the store
	result, err := collection.UpdateOne(
		ctx,
		bson.M{"_id": store.ID},
		bson.M{"$set": store},
	)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrStoreNotFound
	}

	return nil
}

// Delete deletes a store by ID
func (r *MongoStoreRepository) Delete(ctx context.Context, id string) error {
	tenantID, ok := contextfx.GetTenantID(ctx)
	if !ok {
		return contextfx.ErrTenantNotFound
	}

	db := r.client.Database(tenantID)
	collection := db.Collection("stores")

	// Delete the store
	result, err := collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return ErrStoreNotFound
	}

	return nil
}
