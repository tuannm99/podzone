package store

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/fx"

	storeinputport "github.com/tuannm99/podzone/internal/onboarding/store/inputport"
)

var _ storeinputport.Usecase = (*StoreService)(nil)

var (
	ErrStoreNotFound     = errors.New("store not found")
	ErrSubdomainTaken    = errors.New("subdomain is already taken")
	ErrInvalidStatus     = errors.New("invalid store status")
	ErrStoreNotActive    = errors.New("store is not active")
	ErrStoreNotCompleted = errors.New("store onboarding is not completed")
)

// StoreService handles store business logic
type StoreService struct {
	collection storeCollection
}

type storeCollection interface {
	FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) *mongo.SingleResult
	InsertOne(
		ctx context.Context,
		document interface{},
		opts ...*options.InsertOneOptions,
	) (*mongo.InsertOneResult, error)
	Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (*mongo.Cursor, error)
	UpdateOne(
		ctx context.Context,
		filter interface{},
		update interface{},
		opts ...*options.UpdateOptions,
	) (*mongo.UpdateResult, error)
	Indexes() mongo.IndexView
}

// StoreServiceParams contains dependencies for Service
type StoreServiceParams struct {
	fx.In

	MongoClient *mongo.Client `name:"mongo-onboarding"`
}

// NewStoreService creates a new store service
func NewStoreService(params StoreServiceParams) *StoreService {
	collection := params.MongoClient.Database("podzone").Collection("stores")

	// Create indexes
	ctx := context.Background()
	_, _ = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "subdomain", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	_, _ = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "owner_id", Value: 1}},
	})

	return &StoreService{
		collection: collection,
	}
}

// CreateStore creates a new store
func (s *StoreService) CreateStore(ctx context.Context, name, subdomain, ownerID string) (*storeinputport.Store, error) {
	// Check if subdomain is taken
	var existing Store
	err := s.collection.FindOne(ctx, bson.M{"subdomain": subdomain}).Decode(&existing)
	if err == nil {
		return nil, ErrSubdomainTaken
	} else if err != mongo.ErrNoDocuments {
		return nil, err
	}

	store := &Store{
		Name:      name,
		Subdomain: subdomain,
		OwnerID:   ownerID,
		Status:    StoreStatusDraft,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	result, err := s.collection.InsertOne(ctx, store)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, ErrSubdomainTaken
		}
		return nil, err
	}

	store.ID = result.InsertedID.(primitive.ObjectID)
	return toInputPortStore(store), nil
}

// GetStore retrieves a store by ID
func (s *StoreService) GetStore(ctx context.Context, id string) (*storeinputport.Store, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var store Store
	err = s.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&store)
	if err == mongo.ErrNoDocuments {
		return nil, ErrStoreNotFound
	}
	if err != nil {
		return nil, err
	}

	return toInputPortStore(&store), nil
}

// GetStoresByOwner retrieves all stores owned by a user
func (s *StoreService) GetStoresByOwner(ctx context.Context, ownerID string) ([]*storeinputport.Store, error) {
	cursor, err := s.collection.Find(ctx, bson.M{"owner_id": ownerID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var stores []*Store
	if err = cursor.All(ctx, &stores); err != nil {
		return nil, err
	}

	out := make([]*storeinputport.Store, 0, len(stores))
	for _, st := range stores {
		out = append(out, toInputPortStore(st))
	}
	return out, nil
}

// UpdateStoreStatus updates a store's status
func (s *StoreService) UpdateStoreStatus(ctx context.Context, id string, status storeinputport.StoreStatus) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	var store Store
	err = s.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&store)
	if err == mongo.ErrNoDocuments {
		return ErrStoreNotFound
	}
	if err != nil {
		return err
	}

	// Validate status transition
	if !isValidStatusTransition(storeinputport.StoreStatus(store.Status), status) {
		return ErrInvalidStatus
	}

	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	if status == storeinputport.StoreStatusActive {
		now := time.Now()
		update["$set"].(bson.M)["completed_at"] = now
	}

	_, err = s.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	return err
}

// isValidStatusTransition checks if the status transition is valid
func isValidStatusTransition(current, next storeinputport.StoreStatus) bool {
	switch current {
	case storeinputport.StoreStatusDraft:
		return next == storeinputport.StoreStatusPending || next == storeinputport.StoreStatusRejected
	case storeinputport.StoreStatusPending:
		return next == storeinputport.StoreStatusActive || next == storeinputport.StoreStatusRejected
	case storeinputport.StoreStatusActive:
		return next == storeinputport.StoreStatusSuspended
	case storeinputport.StoreStatusRejected:
		return next == storeinputport.StoreStatusPending
	case storeinputport.StoreStatusSuspended:
		return next == storeinputport.StoreStatusActive
	default:
		return false
	}
}

func toInputPortStore(st *Store) *storeinputport.Store {
	if st == nil {
		return nil
	}
	out := &storeinputport.Store{
		ID:        st.ID.Hex(),
		Name:      st.Name,
		Subdomain: st.Subdomain,
		OwnerID:   st.OwnerID,
		Status:    storeinputport.StoreStatus(st.Status),
		CreatedAt: st.CreatedAt,
		UpdatedAt: st.UpdatedAt,
	}
	if st.CompletedAt != nil {
		completedAt := *st.CompletedAt
		out.CompletedAt = &completedAt
	}
	return out
}
