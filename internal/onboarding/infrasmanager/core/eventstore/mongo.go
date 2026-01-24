package eventstore

import (
	"context"
	"time"

	"github.com/tuannm99/podzone/internal/onboarding/infrasmanager/core"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/fx"
)

var _ core.ConnectionStore = (*MongoStore)(nil)

type MongoStore struct {
	db *mongo.Database

	connCol   *mongo.Collection
	eventCol  *mongo.Collection
	outboxCol *mongo.Collection
}

type MongoStoreParams struct {
	fx.In

	Client *mongo.Client `name:"mongo-onboarding"`
	DB     string        `name:"mongo-onboarding-db"`
}

func NewMongoStore(p MongoStoreParams) *MongoStore {
	db := p.Client.Database(p.DB)
	return &MongoStore{
		db:        db,
		connCol:   db.Collection("connections"),
		eventCol:  db.Collection("connection_events"),
		outboxCol: db.Collection("outbox"),
	}
}

func (s *MongoStore) EnsureIndexes(ctx context.Context) error {
	_, err := s.connCol.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "tenant_id", Value: 1}, {Key: "infra_type", Value: 1}, {Key: "name", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "infra_type", Value: 1}, {Key: "status", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "tenant_id", Value: 1}, {Key: "updated_at", Value: -1}},
		},
	})
	if err != nil {
		return err
	}

	// events: unique by "id" (NOT correlation_id)
	_, err = s.eventCol.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "id", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("uniq_event_id"),
		},
		{
			Keys: bson.D{{Key: "correlation_id", Value: 1}, {Key: "created_at", Value: -1}},
		},
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "infra_type", Value: 1},
				{Key: "name", Value: 1},
				{Key: "created_at", Value: -1},
			},
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
	if info.Status == "" {
		info.Status = "active"
	}

	filter := bson.M{
		"tenant_id":  info.TenantID,
		"infra_type": string(info.InfraType),
		"name":       info.Name,
	}

	// IMPORTANT: Fix version bug.
	// Use $setOnInsert version=0 then $inc version=1 => first insert => 1.
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
		if err == mongo.ErrNoDocuments {
			return nil, core.ErrConnectionNotFound
		}
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

func (s *MongoStore) ListConnections(
	tenantID string,
	infraType core.InfraType,
	includeDeleted bool,
	limit, offset int,
) ([]core.ConnectionInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	filter := bson.M{"tenant_id": tenantID}
	if infraType != "" {
		filter["infra_type"] = string(infraType)
	}
	if !includeDeleted {
		filter["deleted_at"] = bson.M{"$eq": nil}
	}

	cur, err := s.connCol.Find(
		ctx,
		filter,
		options.Find().
			SetSort(bson.D{{Key: "updated_at", Value: -1}}).
			SetSkip(int64(offset)).
			SetLimit(int64(limit)),
	)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	out := make([]core.ConnectionInfo, 0, limit)
	for cur.Next(ctx) {
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
		if err := cur.Decode(&doc); err != nil {
			return nil, err
		}
		out = append(out, core.ConnectionInfo{
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
		})
	}
	return out, cur.Err()
}

func (s *MongoStore) ListEvents(
	tenantID string,
	infraType core.InfraType,
	name string,
	correlationID string,
	limit, offset int,
) ([]core.ConnectionEvent, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	filter := bson.M{"tenant_id": tenantID}
	if infraType != "" {
		filter["infra_type"] = string(infraType)
	}
	if name != "" {
		filter["name"] = name
	}
	if correlationID != "" {
		filter["correlation_id"] = correlationID
	}

	cur, err := s.eventCol.Find(
		ctx,
		filter,
		options.Find().
			SetSort(bson.D{{Key: "created_at", Value: -1}}).
			SetSkip(int64(offset)).
			SetLimit(int64(limit)),
	)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	out := make([]core.ConnectionEvent, 0, limit)
	for cur.Next(ctx) {
		var doc struct {
			ID            string                 `bson:"id"`
			CorrelationID string                 `bson:"correlation_id"`
			TenantID      string                 `bson:"tenant_id"`
			InfraType     string                 `bson:"infra_type"`
			Name          string                 `bson:"name"`
			Action        string                 `bson:"action"`
			Status        string                 `bson:"status"`
			Request       map[string]interface{} `bson:"request"`
			Result        map[string]interface{} `bson:"result"`
			Error         string                 `bson:"error"`
			Actor         map[string]string      `bson:"actor"`
			CreatedAt     time.Time              `bson:"created_at"`
		}
		if err := cur.Decode(&doc); err != nil {
			return nil, err
		}
		out = append(out, core.ConnectionEvent{
			ID:            doc.ID,
			CorrelationID: doc.CorrelationID,
			TenantID:      doc.TenantID,
			InfraType:     core.InfraType(doc.InfraType),
			Name:          doc.Name,
			Action:        doc.Action,
			Status:        doc.Status,
			Request:       doc.Request,
			Result:        doc.Result,
			Error:         doc.Error,
			Actor:         doc.Actor,
			CreatedAt:     doc.CreatedAt,
		})
	}
	return out, cur.Err()
}

func (s *MongoStore) AppendEvent(ev core.ConnectionEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if ev.ID == "" {
		return core.ErrInvalidInput
	}
	if ev.CreatedAt.IsZero() {
		ev.CreatedAt = time.Now()
	}
	if ev.Name == "" {
		ev.Name = "default"
	}

	_, err := s.eventCol.InsertOne(ctx, bson.M{
		"id":             ev.ID,
		"correlation_id": ev.CorrelationID,
		"tenant_id":      ev.TenantID,
		"infra_type":     string(ev.InfraType),
		"name":           ev.Name,
		"action":         ev.Action,
		"status":         ev.Status,
		"request":        ev.Request,
		"result":         ev.Result,
		"error":          ev.Error,
		"actor":          ev.Actor,
		"created_at":     ev.CreatedAt,
	})
	if mongo.IsDuplicateKeyError(err) {
		return nil
	}
	return err
}

func (s *MongoStore) EnqueueOutbox(msg core.OutboxMessage) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if msg.EventID == "" {
		return core.ErrInvalidInput
	}

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
				"correlation_id": msg.CorrelationID,
				"topic":          msg.Topic,
				"payload":        msg.Payload,
				"tenant_id":      msg.TenantID,
				"infra_type":     string(msg.InfraType),
				"name":           msg.Name,
				"status":         msg.Status,
				"retry_count":    msg.RetryCount,
				"next_retry":     msg.NextRetry,
				"updated_at":     msg.UpdatedAt,
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
			EventID       string                 `bson:"event_id"`
			CorrelationID string                 `bson:"correlation_id"`
			Topic         string                 `bson:"topic"`
			Payload       map[string]interface{} `bson:"payload"`
			TenantID      string                 `bson:"tenant_id"`
			InfraType     string                 `bson:"infra_type"`
			Name          string                 `bson:"name"`
			Status        string                 `bson:"status"`
			RetryCount    int                    `bson:"retry_count"`
			NextRetry     time.Time              `bson:"next_retry"`
			CreatedAt     time.Time              `bson:"created_at"`
			UpdatedAt     time.Time              `bson:"updated_at"`
		}
		if err := cur.Decode(&doc); err != nil {
			return nil, err
		}
		out = append(out, core.OutboxMessage{
			EventID:       doc.EventID,
			CorrelationID: doc.CorrelationID,
			Topic:         doc.Topic,
			Payload:       doc.Payload,
			TenantID:      doc.TenantID,
			InfraType:     core.InfraType(doc.InfraType),
			Name:          doc.Name,
			Status:        doc.Status,
			RetryCount:    doc.RetryCount,
			NextRetry:     doc.NextRetry,
			CreatedAt:     doc.CreatedAt,
			UpdatedAt:     doc.UpdatedAt,
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
