package store

import (
	"context"
	"errors"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type fakeCollection struct {
	findOneResult   *mongo.SingleResult
	insertOneResult *mongo.InsertOneResult
	insertOneErr    error
}

func (f *fakeCollection) FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) *mongo.SingleResult {
	return f.findOneResult
}

func (f *fakeCollection) InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	return f.insertOneResult, f.insertOneErr
}

func (f *fakeCollection) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (*mongo.Cursor, error) {
	return nil, errors.New("unexpected call to Find")
}

func (f *fakeCollection) UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return nil, errors.New("unexpected call to UpdateOne")
}

func (f *fakeCollection) Indexes() mongo.IndexView {
	return mongo.IndexView{}
}

func TestCreateStore_ReturnsErrSubdomainTaken_WhenExistingStoreFound(t *testing.T) {
	existing := Store{
		ID:        primitive.NewObjectID(),
		Name:      "Existing",
		Subdomain: "taken",
		OwnerID:   "owner-1",
		Status:    StoreStatusActive,
	}

	service := &StoreService{
		collection: &fakeCollection{
			findOneResult: mongo.NewSingleResultFromDocument(existing, nil, nil),
		},
	}

	_, err := service.CreateStore(context.Background(), "New", "taken", "owner-2")
	if !errors.Is(err, ErrSubdomainTaken) {
		t.Fatalf("expected ErrSubdomainTaken, got %v", err)
	}
}

func TestCreateStore_ReturnsErrSubdomainTaken_WhenInsertDuplicateKey(t *testing.T) {
	service := &StoreService{
		collection: &fakeCollection{
			findOneResult: mongo.NewSingleResultFromDocument(bson.D{}, mongo.ErrNoDocuments, nil),
			insertOneErr: mongo.WriteException{
				WriteErrors: []mongo.WriteError{{Code: 11000}},
			},
		},
	}

	_, err := service.CreateStore(context.Background(), "New", "taken", "owner-2")
	if !errors.Is(err, ErrSubdomainTaken) {
		t.Fatalf("expected ErrSubdomainTaken, got %v", err)
	}
}
