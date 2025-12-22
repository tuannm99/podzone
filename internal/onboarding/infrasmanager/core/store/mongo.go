package store

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tuannm99/podzone/internal/onboarding/infrasmanager/core"
)

type MongoStore struct {
	db *mongo.Database

	connCol   *mongo.Collection
	eventCol  *mongo.Collection
	outboxCol *mongo.Collection
}

func NewMongoStore(db *mongo.Database) *MongoStore {
	return &MongoStore{
		db:        db,
		connCol:   db.Collection("connections"),
		eventCol:  db.Collection("connection_events"),
		outboxCol: db.Collection("outbox"),
	}
}

func (s *MongoStore) EnsureIndexes(ctx context.Context) error {
	// connections: unique (tenant_id, infra_type, name)
	_, err := s.connCol.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "tenant_id", Value: 1}, {Key: "infra_type", Value: 1}, {Key: "name", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("uniq_tenant_infra_name"),
		},
		{
			Keys:    bson.D{{Key: "infra_type", Value: 1}, {Key: "status", Value: 1}},
			Options: options.Index().SetName("infra_status"),
		},
	})
	if err != nil {
		return err
	}

	// events: unique event_id
	_, err = s.eventCol.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "event_id", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("uniq_event_id"),
		},
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "infra_type", Value: 1},
				{Key: "created_at", Value: -1},
			},
			Options: options.Index().SetName("tenant_infra_time"),
		},
	})
	if err != nil {
		return err
	}

	// outbox: query due
	_, err = s.outboxCol.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "status", Value: 1}, {Key: "next_retry", Value: 1}},
			Options: options.Index().SetName("status_next_retry"),
		},
		{
			Keys:    bson.D{{Key: "event_id", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("uniq_outbox_event"),
		},
	})
	return err
}

func (s *MongoStore) Upsert(info core.ConnectionInfo) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if info.Name == "" {
		info.Name = "default"
	}
	now := time.Now()
	if info.CreatedAt.IsZero() {
		info.CreatedAt = now
	}
	info.UpdatedAt = now

	filter := bson.M{
		"tenant_id":  info.TenantID,
		"infra_type": string(info.InfraType),
		"name":       info.Name,
	}
	update := bson.M{
		"$set": bson.M{
			"endpoint":   info.Endpoint,
			"secret_ref": info.SecretRef,
			"status":     info.Status,
			"meta":       info.Meta,
			"config":     info.Config,
			"updated_at": info.UpdatedAt,
			"deleted_at": nil,
		},
		"$setOnInsert": bson.M{
			"created_at": info.CreatedAt,
			"version":    int64(1),
		},
		"$inc": bson.M{"version": int64(1)},
	}

	_, err := s.connCol.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	return err
}

func (s *MongoStore) SoftDelete(tenantID string, infraType core.InfraType, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if name == "" {
		name = "default"
	}
	now := time.Now()

	filter := bson.M{"tenant_id": tenantID, "infra_type": string(infraType), "name": name}
	update := bson.M{
		"$set": bson.M{
			"status":     "deleted",
			"deleted_at": now,
			"updated_at": now,
		},
		"$inc": bson.M{"version": int64(1)},
	}
	_, err := s.connCol.UpdateOne(ctx, filter, update)
	return err
}

func (s *MongoStore) Get(tenantID string, infraType core.InfraType, name string) (*core.ConnectionInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if name == "" {
		name = "default"
	}

	var doc struct {
		TenantID  string                 `bson:"tenant_id"`
		InfraType string                 `bson:"infra_type"`
		Name      string                 `bson:"name"`
		Endpoint  string                 `bson:"endpoint"`
		SecretRef string                 `bson:"secret_ref"`
		Status    string                 `bson:"status"`
		Version   int64                  `bson:"version"`
		Meta      map[string]string      `bson:"meta"`
		Config    map[string]interface{} `bson:"config"`
		CreatedAt time.Time              `bson:"created_at"`
		UpdatedAt time.Time              `bson:"updated_at"`
		DeletedAt *time.Time             `bson:"deleted_at,omitempty"`
	}

	err := s.connCol.FindOne(ctx, bson.M{
		"tenant_id": tenantID, "infra_type": string(infraType), "name": name,
	}).Decode(&doc)
	if err != nil {
		return nil, err
	}

	return &core.ConnectionInfo{
		TenantID:  doc.TenantID,
		InfraType: core.InfraType(doc.InfraType),
		Name:      doc.Name,
		Endpoint:  doc.Endpoint,
		SecretRef: doc.SecretRef,
		Status:    doc.Status,
		Version:   doc.Version,
		Meta:      doc.Meta,
		Config:    doc.Config,
		CreatedAt: doc.CreatedAt,
		UpdatedAt: doc.UpdatedAt,
		DeletedAt: doc.DeletedAt,
	}, nil
}

func (s *MongoStore) AppendEvent(ev core.ConnectionEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if ev.CreatedAt.IsZero() {
		ev.CreatedAt = time.Now()
	}
	if ev.Name == "" {
		ev.Name = "default"
	}

	_, err := s.eventCol.InsertOne(ctx, bson.M{
		"event_id":   ev.EventID,
		"tenant_id":  ev.TenantID,
		"infra_type": string(ev.InfraType),
		"name":       ev.Name,
		"action":     ev.Action,
		"status":     ev.Status,
		"request":    ev.Request,
		"result":     ev.Result,
		"error":      ev.Error,
		"actor":      ev.Actor,
		"created_at": ev.CreatedAt,
	})
	// If duplicate event_id, ignore to support idempotency.
	if mongo.IsDuplicateKeyError(err) {
		return nil
	}
	return err
}

func (s *MongoStore) EnqueueOutbox(msg core.OutboxMessage) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()
	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = now
	}
	msg.UpdatedAt = now
	if msg.Status == "" {
		msg.Status = "pending"
	}
	if msg.NextRetry.IsZero() {
		msg.NextRetry = now
	}

	_, err := s.outboxCol.UpdateOne(
		ctx,
		bson.M{"event_id": msg.EventID},
		bson.M{
			"$setOnInsert": bson.M{
				"event_id":   msg.EventID,
				"created_at": msg.CreatedAt,
			},
			"$set": bson.M{
				"topic":       msg.Topic,
				"payload":     msg.Payload,
				"status":      msg.Status,
				"retry_count": msg.RetryCount,
				"next_retry":  msg.NextRetry,
				"updated_at":  msg.UpdatedAt,
			},
		},
		options.Update().SetUpsert(true),
	)
	return err
}

func (s *MongoStore) FindDueOutbox(limit int) ([]core.OutboxMessage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if limit <= 0 {
		limit = 50
	}

	cur, err := s.outboxCol.Find(ctx, bson.M{
		"status":     bson.M{"$in": []string{"pending", "failed"}},
		"next_retry": bson.M{"$lte": time.Now()},
	}, options.Find().SetSort(bson.D{{Key: "next_retry", Value: 1}}).SetLimit(int64(limit)))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var out []core.OutboxMessage
	for cur.Next(ctx) {
		var doc struct {
			EventID    string                 `bson:"event_id"`
			Topic      string                 `bson:"topic"`
			Payload    map[string]interface{} `bson:"payload"`
			Status     string                 `bson:"status"`
			RetryCount int                    `bson:"retry_count"`
			NextRetry  time.Time              `bson:"next_retry"`
			CreatedAt  time.Time              `bson:"created_at"`
			UpdatedAt  time.Time              `bson:"updated_at"`
		}
		if err := cur.Decode(&doc); err != nil {
			return nil, err
		}
		out = append(out, core.OutboxMessage{
			EventID:    doc.EventID,
			Topic:      doc.Topic,
			Payload:    doc.Payload,
			Status:     doc.Status,
			RetryCount: doc.RetryCount,
			NextRetry:  doc.NextRetry,
			CreatedAt:  doc.CreatedAt,
			UpdatedAt:  doc.UpdatedAt,
		})
	}
	return out, cur.Err()
}

func (s *MongoStore) MarkOutboxDone(eventID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.outboxCol.UpdateOne(ctx, bson.M{"event_id": eventID}, bson.M{
		"$set": bson.M{"status": "done", "updated_at": time.Now()},
	})
	return err
}

func (s *MongoStore) MarkOutboxFailed(eventID string, nextRetry time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.outboxCol.UpdateOne(ctx, bson.M{"event_id": eventID}, bson.M{
		"$set": bson.M{"status": "failed", "next_retry": nextRetry, "updated_at": time.Now()},
		"$inc": bson.M{"retry_count": 1},
	})
	return err
}
