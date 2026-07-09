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
	"github.com/tuannm99/podzone/pkg/collection"
)

var _ storeoutputport.StoreRepository = (*MongoRepository)(nil)

type MongoRepository struct {
	collection    *mongo.Collection
	transitionCol *mongo.Collection
}

type Params struct {
	fx.In

	MongoClient *mongo.Client `name:"mongo-onboarding"`
	DB          string        `name:"mongo-onboarding-db"`
}

func New(p Params) *MongoRepository {
	db := p.MongoClient.Database(p.DB)
	return &MongoRepository{
		collection:    db.Collection("store_requests"),
		transitionCol: db.Collection("store_request_transitions"),
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
	if err != nil {
		return err
	}
	_, err = r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "status", Value: 1},
			{Key: "lease_until", Value: 1},
			{Key: "updated_at", Value: 1},
		},
	})
	if err != nil {
		return err
	}
	_, err = r.transitionCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "request_id", Value: 1}, {Key: "created_at", Value: -1}},
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

func (r *MongoRepository) Create(
	ctx context.Context,
	request storeentity.StoreRequest,
) (*storeentity.StoreRequest, error) {
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

func (r *MongoRepository) ListPage(
	ctx context.Context,
	workspaceID string,
	query collection.Query,
) (collection.Page[storeentity.StoreRequest], error) {
	normalized, filter, sort, err := buildStoreRequestCollection(workspaceID, query)
	if err != nil {
		return collection.Page[storeentity.StoreRequest]{}, err
	}
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return collection.Page[storeentity.StoreRequest]{}, err
	}

	cursor, err := r.collection.Find(
		ctx,
		filter,
		options.Find().
			SetSort(sort).
			SetSkip(int64(normalized.Offset())).
			SetLimit(int64(normalized.PageSize)),
	)
	if err != nil {
		return collection.Page[storeentity.StoreRequest]{}, err
	}
	defer cursor.Close(ctx)

	var requests []storeentity.StoreRequest
	if err = cursor.All(ctx, &requests); err != nil {
		return collection.Page[storeentity.StoreRequest]{}, err
	}
	return collection.NewPage(requests, total, normalized), nil
}

func (r *MongoRepository) ListTransitions(
	ctx context.Context,
	requestID string,
	query collection.Query,
) (collection.Page[storeentity.StoreRequestTransition], error) {
	normalized, filter, sort, err := buildStoreTransitionCollection(requestID, query)
	if err != nil {
		return collection.Page[storeentity.StoreRequestTransition]{}, err
	}
	total, err := r.transitionCol.CountDocuments(ctx, filter)
	if err != nil {
		return collection.Page[storeentity.StoreRequestTransition]{}, err
	}
	cursor, err := r.transitionCol.Find(
		ctx,
		filter,
		options.Find().
			SetSort(sort).
			SetSkip(int64(normalized.Offset())).
			SetLimit(int64(normalized.PageSize)),
	)
	if err != nil {
		return collection.Page[storeentity.StoreRequestTransition]{}, err
	}
	defer cursor.Close(ctx)

	var transitions []storeentity.StoreRequestTransition
	if err := cursor.All(ctx, &transitions); err != nil {
		return collection.Page[storeentity.StoreRequestTransition]{}, err
	}
	return collection.NewPage(transitions, total, normalized), nil
}

func (r *MongoRepository) ClaimNextQueued(
	ctx context.Context,
	leaseOwner string,
	leaseTTL time.Duration,
) (*storeentity.StoreRequest, error) {
	return r.claimNext(
		ctx,
		[]storeentity.RequestStatus{storeentity.RequestStatusQueued},
		storeentity.RequestStatusPlanning,
		leaseOwner,
		leaseTTL,
	)
}

func (r *MongoRepository) ClaimNextProvisioning(
	ctx context.Context,
	leaseOwner string,
	leaseTTL time.Duration,
) (*storeentity.StoreRequest, error) {
	return r.claimNext(
		ctx,
		[]storeentity.RequestStatus{storeentity.RequestStatusProvisioning},
		storeentity.RequestStatusProvisioning,
		leaseOwner,
		leaseTTL,
	)
}

func (r *MongoRepository) claimNext(
	ctx context.Context,
	current []storeentity.RequestStatus,
	next storeentity.RequestStatus,
	leaseOwner string,
	leaseTTL time.Duration,
) (*storeentity.StoreRequest, error) {
	now := time.Now().UTC()
	leaseUntil := now.Add(leaseTTL)
	var request storeentity.StoreRequest
	err := r.collection.FindOneAndUpdate(
		ctx,
		bson.M{
			"status": bson.M{"$in": current},
			"$or": bson.A{
				bson.M{"lease_until": bson.M{"$exists": false}},
				bson.M{"lease_until": nil},
				bson.M{"lease_until": bson.M{"$lte": now}},
			},
		},
		bson.M{
			"$set": bson.M{
				"status":      next,
				"last_error":  "",
				"lease_owner": leaseOwner,
				"lease_until": leaseUntil,
				"updated_at":  now,
			},
			"$inc": bson.M{"attempt": 1},
		},
		options.FindOneAndUpdate().
			SetSort(bson.D{{Key: "updated_at", Value: 1}}).
			SetReturnDocument(options.Before),
	).Decode(&request)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &request, nil
}

func (r *MongoRepository) ReleaseLease(ctx context.Context, id string, leaseOwner string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID, "lease_owner": leaseOwner},
		bson.M{"$unset": bson.M{"lease_owner": "", "lease_until": ""}},
	)
	return err
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
		"$unset": bson.M{
			"lease_owner": "",
			"lease_until": "",
		},
	}
	if status == storeentity.RequestStatusQueued {
		update["$set"].(bson.M)["approved_at"] = now
	}
	if isTerminalStatus(status) {
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
			"$unset": bson.M{"lease_owner": "", "lease_until": ""},
		},
	)
	return err
}

func (r *MongoRepository) MarkFailed(ctx context.Context, id string, reason string) error {
	return r.MarkBlocked(ctx, id, storeentity.RequestStatusFailed, reason)
}

func (r *MongoRepository) MarkBlocked(
	ctx context.Context,
	id string,
	status storeentity.RequestStatus,
	reason string,
) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"last_error": reason,
			"updated_at": now,
		},
		"$unset": bson.M{
			"lease_owner": "",
			"lease_until": "",
		},
	}
	if isTerminalStatus(status) {
		update["$set"].(bson.M)["completed_at"] = now
	}
	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		update,
	)
	return err
}

func (r *MongoRepository) RecordTransition(
	ctx context.Context,
	transition storeentity.StoreRequestTransition,
) error {
	if transition.RequestID == "" {
		return nil
	}
	if transition.CreatedAt.IsZero() {
		transition.CreatedAt = time.Now().UTC()
	}
	_, err := r.transitionCol.InsertOne(ctx, transition)
	return err
}

func isTerminalStatus(status storeentity.RequestStatus) bool {
	switch status {
	case storeentity.RequestStatusReady,
		storeentity.RequestStatusFailed,
		storeentity.RequestStatusFailedNonRetryable,
		storeentity.RequestStatusRejected,
		storeentity.RequestStatusArchived,
		storeentity.RequestStatusCancelled:
		return true
	default:
		return false
	}
}
