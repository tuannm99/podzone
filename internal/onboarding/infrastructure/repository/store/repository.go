package repository

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/fx"

	storeentity "github.com/tuannm99/podzone/internal/onboarding/domain/store/entity"
	storeoutputport "github.com/tuannm99/podzone/internal/onboarding/domain/store/outputport"
)

var _ storeoutputport.StoreRepository = (*MongoRepository)(nil)

type MongoRepository struct {
	collection *mongo.Collection
}

type Params struct {
	fx.In

	MongoClient *mongo.Client `name:"mongo-onboarding"`
}

func New(p Params) *MongoRepository {
	return &MongoRepository{
		collection: p.MongoClient.Database("podzone").Collection("store_requests"),
	}
}

func (r *MongoRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "subdomain", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return err
	}
	_, err = r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "workspace_id", Value: 1}, {Key: "status", Value: 1}},
	})
	if err != nil {
		return err
	}
	_, err = r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "workspace_id", Value: 1}, {Key: "updated_at", Value: -1}},
	})
	return err
}

func (r *MongoRepository) FindBySubdomain(ctx context.Context, subdomain string) (*storeentity.StoreRequest, error) {
	var request storeentity.StoreRequest
	err := r.collection.FindOne(ctx, bson.M{"subdomain": subdomain}).Decode(&request)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &request, nil
}

func (r *MongoRepository) Create(ctx context.Context, request storeentity.StoreRequest) (*storeentity.StoreRequest, error) {
	result, err := r.collection.InsertOne(ctx, request)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, errors.New("duplicate store request")
		}
		return nil, err
	}
	request.ID = result.InsertedID.(primitive.ObjectID)
	return &request, nil
}

func (r *MongoRepository) FindByID(ctx context.Context, id string) (*storeentity.StoreRequest, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var request storeentity.StoreRequest
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&request)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &request, nil
}

func (r *MongoRepository) List(ctx context.Context, workspaceID string) ([]storeentity.StoreRequest, error) {
	filter := bson.M{}
	if workspaceID != "" {
		filter["workspace_id"] = workspaceID
	}

	cursor, err := r.collection.Find(
		ctx,
		filter,
		options.Find().SetSort(bson.D{{Key: "updated_at", Value: -1}}),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var requests []storeentity.StoreRequest
	if err = cursor.All(ctx, &requests); err != nil {
		return nil, err
	}
	return requests, nil
}

func (r *MongoRepository) ClaimNextQueued(ctx context.Context) (*storeentity.StoreRequest, error) {
	now := time.Now().UTC()
	var request storeentity.StoreRequest
	err := r.collection.FindOneAndUpdate(
		ctx,
		bson.M{"status": storeentity.RequestStatusQueued},
		bson.M{
			"$set": bson.M{
				"status":     storeentity.RequestStatusProvisioning,
				"last_error": "",
				"updated_at": now,
			},
		},
		options.FindOneAndUpdate().
			SetSort(bson.D{{Key: "updated_at", Value: 1}}).
			SetReturnDocument(options.After),
	).Decode(&request)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &request, nil
}

func (r *MongoRepository) UpdateStatus(ctx context.Context, id string, status storeentity.RequestStatus) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": now,
		},
	}
	if status == storeentity.RequestStatusQueued {
		update["$set"].(bson.M)["approved_at"] = now
	}
	if status == storeentity.RequestStatusReady || status == storeentity.RequestStatusFailed {
		update["$set"].(bson.M)["completed_at"] = now
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	return err
}

func (r *MongoRepository) MarkReady(ctx context.Context, id string, storeID string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	storeObjectID, err := primitive.ObjectIDFromHex(storeID)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{
			"$set": bson.M{
				"status":       storeentity.RequestStatusReady,
				"store_id":     storeObjectID,
				"last_error":   "",
				"updated_at":   now,
				"completed_at": now,
			},
		},
	)
	return err
}

func (r *MongoRepository) MarkFailed(ctx context.Context, id string, reason string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{
			"$set": bson.M{
				"status":       storeentity.RequestStatusFailed,
				"last_error":   reason,
				"updated_at":   now,
				"completed_at": now,
			},
		},
	)
	return err
}
