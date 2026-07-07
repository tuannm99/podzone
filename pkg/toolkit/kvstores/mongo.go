package kvstores

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/tuannm99/podzone/pkg/pdlog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const mongoOperationTimeout = 5 * time.Second

var _ KVStore = (*MongoStore)(nil)

type mongoEntry struct {
	Key       string    `bson:"_id"`
	Value     []byte    `bson:"value"`
	UpdatedAt time.Time `bson:"updated_at"`
}

// MongoStore persists runtime routing projections as key/value documents.
type MongoStore struct {
	collection *mongo.Collection
	logger     pdlog.Logger
}

func NewMongoStore(
	client *mongo.Client,
	database string,
	collection string,
	logger pdlog.Logger,
) (*MongoStore, error) {
	if client == nil {
		return nil, errors.New("mongo kv store client is nil")
	}
	if database == "" {
		return nil, errors.New("mongo kv store database is required")
	}
	if collection == "" {
		return nil, errors.New("mongo kv store collection is required")
	}
	return &MongoStore{
		collection: client.Database(database).Collection(collection),
		logger:     logger,
	}, nil
}

func (s *MongoStore) Get(ctx context.Context, path string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, mongoOperationTimeout)
	defer cancel()

	var entry mongoEntry
	if err := s.collection.FindOne(ctx, bson.M{"_id": path}).Decode(&entry); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("%w: %s", ErrKeyNotFound, path)
		}
		return nil, fmt.Errorf("get kv key %q: %w", path, err)
	}
	s.logger.Debug("KV loaded", "key", path)
	return append([]byte(nil), entry.Value...), nil
}

func (s *MongoStore) GetKVs(ctx context.Context, prefix string) (map[string][]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, mongoOperationTimeout)
	defer cancel()

	cursor, err := s.collection.Find(
		ctx,
		bson.M{"_id": bson.M{"$regex": "^" + regexp.QuoteMeta(prefix)}},
		options.Find().SetSort(bson.D{{Key: "_id", Value: 1}}),
	)
	if err != nil {
		return nil, fmt.Errorf("list kv prefix %q: %w", prefix, err)
	}
	defer cursor.Close(ctx)

	entries := make([]mongoEntry, 0)
	if err := cursor.All(ctx, &entries); err != nil {
		return nil, fmt.Errorf("decode kv prefix %q: %w", prefix, err)
	}
	result := make(map[string][]byte, len(entries))
	for _, entry := range entries {
		result[entry.Key] = append([]byte(nil), entry.Value...)
	}
	s.logger.Debug("KV prefix loaded", "prefix", prefix, "count", len(result))
	return result, nil
}

func (s *MongoStore) Put(ctx context.Context, path string, value []byte) error {
	ctx, cancel := context.WithTimeout(ctx, mongoOperationTimeout)
	defer cancel()

	_, err := s.collection.UpdateOne(
		ctx,
		bson.M{"_id": path},
		bson.M{"$set": bson.M{
			"value":      append([]byte(nil), value...),
			"updated_at": time.Now().UTC(),
		}},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		return fmt.Errorf("put kv key %q: %w", path, err)
	}
	s.logger.Debug("KV stored", "key", path, "bytes", len(value))
	return nil
}

func (s *MongoStore) Del(ctx context.Context, path string) error {
	ctx, cancel := context.WithTimeout(ctx, mongoOperationTimeout)
	defer cancel()

	if _, err := s.collection.DeleteOne(ctx, bson.M{"_id": path}); err != nil {
		return fmt.Errorf("delete kv key %q: %w", path, err)
	}
	s.logger.Debug("KV deleted", "key", path)
	return nil
}
