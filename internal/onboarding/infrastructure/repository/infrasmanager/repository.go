package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/entity"
	storeoutputport "github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/outputport"
	"github.com/tuannm99/podzone/pkg/messaging"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/fx"
)

var _ storeoutputport.ConnectionStore = (*MongoStore)(nil)
var _ storeoutputport.PlacementRepository = (*MongoStore)(nil)
var _ messaging.OutboxStore = (*MongoStore)(nil)

type MongoStore struct {
	db *mongo.Database

	connCol   *mongo.Collection
	eventCol  *mongo.Collection
	outboxCol *mongo.Collection
	placeCol  *mongo.Collection
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
		outboxCol: db.Collection("connection_outbox"),
		placeCol:  db.Collection("placement_allocations"),
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
			Options: options.Index().SetName("connection_outbox_status_next_retry"),
		},
		{
			Keys:    bson.D{{Key: "event_id", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("uniq_connection_outbox_event"),
		},
	})
	if err != nil {
		return err
	}
	_, err = s.placeCol.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "tenant_id", Value: 1}, {Key: "store_id", Value: 1}},
			Options: options.Index().
				SetUnique(true).
				SetName("uniq_placement_allocation_tenant_store"),
		},
		{
			Keys:    bson.D{{Key: "status", Value: 1}, {Key: "updated_at", Value: -1}},
			Options: options.Index().SetName("placement_allocation_status_updated"),
		},
	})
	return err
}

func (s *MongoStore) Upsert(ctx context.Context, info entity.ConnectionInfo) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
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

func (s *MongoStore) SoftDelete(ctx context.Context, tenantID string, infraType entity.InfraType, name string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
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

func (s *MongoStore) Get(
	ctx context.Context,
	tenantID string,
	infraType entity.InfraType,
	name string,
) (*entity.ConnectionInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
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
			return nil, entity.ErrConnectionNotFound
		}
		return nil, err
	}

	return &entity.ConnectionInfo{
		TenantID:  doc.TenantID,
		InfraType: entity.InfraType(doc.InfraType),
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
	ctx context.Context,
	tenantID string,
	infraType entity.InfraType,
	includeDeleted bool,
	limit, offset int,
) ([]entity.ConnectionInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
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

	out := make([]entity.ConnectionInfo, 0, limit)
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
		out = append(out, entity.ConnectionInfo{
			TenantID:  doc.TenantID,
			InfraType: entity.InfraType(doc.InfraType),
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
	ctx context.Context,
	tenantID string,
	infraType entity.InfraType,
	name string,
	correlationID string,
	limit, offset int,
) ([]entity.ConnectionEvent, error) {
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
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

	out := make([]entity.ConnectionEvent, 0, limit)
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
		out = append(out, entity.ConnectionEvent{
			ID:            doc.ID,
			CorrelationID: doc.CorrelationID,
			TenantID:      doc.TenantID,
			InfraType:     entity.InfraType(doc.InfraType),
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

func (s *MongoStore) AppendEvent(ctx context.Context, ev entity.ConnectionEvent) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if ev.ID == "" {
		return entity.ErrInvalidInput
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

func (s *MongoStore) EnqueueOutbox(ctx context.Context, msg entity.OutboxMessage) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if msg.EventID == "" {
		return entity.ErrInvalidInput
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

func (s *MongoStore) Append(ctx context.Context, tx messaging.Tx, record messaging.OutboxRecord) error {
	_ = tx
	msg, err := outboxRecordToCore(record)
	if err != nil {
		return err
	}
	return s.EnqueueOutbox(ctx, msg)
}

func (s *MongoStore) ListPending(ctx context.Context, limit int) ([]messaging.OutboxRecord, error) {
	msgs, err := s.FindDueOutbox(ctx, limit)
	if err != nil {
		return nil, err
	}
	out := make([]messaging.OutboxRecord, 0, len(msgs))
	for _, msg := range msgs {
		payload, err := json.Marshal(msg.Payload)
		if err != nil {
			return nil, fmt.Errorf("marshal onboarding outbox payload: %w", err)
		}
		out = append(out, messaging.OutboxRecord{
			ID:            msg.EventID,
			Topic:         msg.Topic,
			MessageKey:    outboxMessageKey(msg),
			Envelope:      outboxMessageEnvelope(msg, payload),
			Status:        msg.Status,
			Attempts:      msg.RetryCount,
			NextAttemptAt: msg.NextRetry,
			CreatedAt:     msg.CreatedAt,
			UpdatedAt:     msg.UpdatedAt,
		})
	}
	return out, nil
}

func (s *MongoStore) MarkPublished(ctx context.Context, ids []string, publishedAt time.Time) error {
	_ = publishedAt
	for _, id := range ids {
		if err := s.MarkOutboxDone(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

func (s *MongoStore) MarkFailed(ctx context.Context, id string, errText string, nextAttemptAt time.Time) error {
	_ = errText
	return s.MarkOutboxFailed(ctx, id, nextAttemptAt)
}

func (s *MongoStore) FindDueOutbox(ctx context.Context, limit int) ([]entity.OutboxMessage, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
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

	var out []entity.OutboxMessage
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
		out = append(out, entity.OutboxMessage{
			EventID:       doc.EventID,
			CorrelationID: doc.CorrelationID,
			Topic:         doc.Topic,
			Payload:       doc.Payload,
			TenantID:      doc.TenantID,
			InfraType:     entity.InfraType(doc.InfraType),
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

func (s *MongoStore) MarkOutboxDone(ctx context.Context, eventID string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := s.outboxCol.UpdateOne(ctx, bson.M{"event_id": eventID}, bson.M{
		"$set": bson.M{"status": "done", "updated_at": time.Now()},
	})
	return err
}

func (s *MongoStore) MarkOutboxFailed(ctx context.Context, eventID string, nextRetry time.Time) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := s.outboxCol.UpdateOne(ctx, bson.M{"event_id": eventID}, bson.M{
		"$set": bson.M{"status": "failed", "next_retry": nextRetry, "updated_at": time.Now()},
		"$inc": bson.M{"retry_count": 1},
	})
	return err
}

func (s *MongoStore) GetPlacementAllocation(
	ctx context.Context,
	tenantID string,
	storeID string,
) (*entity.PlacementAllocation, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var doc placementAllocationDoc
	err := s.placeCol.FindOne(ctx, bson.M{"tenant_id": tenantID, "store_id": storeID}).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	out := doc.toEntity()
	return &out, nil
}

func (s *MongoStore) SavePlacementAllocation(
	ctx context.Context,
	allocation entity.PlacementAllocation,
) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	doc := placementAllocationFromEntity(allocation)
	_, err := s.placeCol.UpdateOne(
		ctx,
		bson.M{"tenant_id": allocation.TenantID, "store_id": allocation.StoreID},
		bson.M{
			"$set": doc,
			"$setOnInsert": bson.M{
				"created_at": allocation.CreatedAt,
			},
		},
		options.Update().SetUpsert(true),
	)
	return err
}

type placementAllocationDoc struct {
	ID           string            `bson:"id"`
	RequestID    string            `bson:"request_id"`
	TenantID     string            `bson:"tenant_id"`
	StoreID      string            `bson:"store_id"`
	Runtime      string            `bson:"runtime"`
	ClusterName  string            `bson:"cluster_name"`
	Mode         string            `bson:"mode"`
	DBName       string            `bson:"db_name"`
	SchemaName   string            `bson:"schema_name"`
	Endpoint     string            `bson:"endpoint"`
	SecretRef    string            `bson:"secret_ref"`
	Status       string            `bson:"status"`
	ProviderMeta map[string]string `bson:"provider_meta"`
	CreatedAt    time.Time         `bson:"created_at"`
	UpdatedAt    time.Time         `bson:"updated_at"`
}

func placementAllocationFromEntity(allocation entity.PlacementAllocation) placementAllocationDoc {
	return placementAllocationDoc{
		ID:           allocation.ID,
		RequestID:    allocation.RequestID,
		TenantID:     allocation.TenantID,
		StoreID:      allocation.StoreID,
		Runtime:      string(allocation.Runtime),
		ClusterName:  allocation.ClusterName,
		Mode:         allocation.Mode,
		DBName:       allocation.DBName,
		SchemaName:   allocation.SchemaName,
		Endpoint:     allocation.Endpoint,
		SecretRef:    allocation.SecretRef,
		Status:       allocation.Status,
		ProviderMeta: allocation.ProviderMeta,
		CreatedAt:    allocation.CreatedAt,
		UpdatedAt:    allocation.UpdatedAt,
	}
}

func (d placementAllocationDoc) toEntity() entity.PlacementAllocation {
	return entity.PlacementAllocation{
		ID:           d.ID,
		RequestID:    d.RequestID,
		TenantID:     d.TenantID,
		StoreID:      d.StoreID,
		Runtime:      entity.PlacementRuntime(d.Runtime),
		ClusterName:  d.ClusterName,
		Mode:         d.Mode,
		DBName:       d.DBName,
		SchemaName:   d.SchemaName,
		Endpoint:     d.Endpoint,
		SecretRef:    d.SecretRef,
		Status:       d.Status,
		ProviderMeta: d.ProviderMeta,
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    d.UpdatedAt,
	}
}

func outboxRecordToCore(record messaging.OutboxRecord) (entity.OutboxMessage, error) {
	var payload map[string]interface{}
	if len(record.Envelope.Payload) > 0 {
		if err := json.Unmarshal(record.Envelope.Payload, &payload); err != nil {
			return entity.OutboxMessage{}, fmt.Errorf("unmarshal onboarding outbox payload: %w", err)
		}
	} else {
		payload = map[string]interface{}{}
	}
	return entity.OutboxMessage{
		EventID:       record.ID,
		CorrelationID: record.Envelope.CorrelationID,
		Topic:         record.Envelope.Type,
		Payload:       payload,
		TenantID:      record.Envelope.TenantID,
		InfraType:     entity.InfraType(record.Envelope.Headers["infra_type"]),
		Name:          record.Envelope.EntityID,
		Status:        record.Status,
		RetryCount:    record.Attempts,
		NextRetry:     record.NextAttemptAt,
		CreatedAt:     record.CreatedAt,
		UpdatedAt:     record.UpdatedAt,
	}, nil
}

func outboxMessageEnvelope(msg entity.OutboxMessage, payload []byte) messaging.Envelope {
	return messaging.Envelope{
		ID:            msg.EventID,
		Type:          msg.Topic,
		Source:        "onboarding",
		TenantID:      msg.TenantID,
		EntityID:      msg.Name,
		CorrelationID: msg.CorrelationID,
		OccurredAt:    msg.CreatedAt,
		SchemaVersion: 1,
		Headers: map[string]string{
			"infra_type": string(msg.InfraType),
		},
		Payload: payload,
	}
}

func outboxMessageKey(msg entity.OutboxMessage) string {
	if msg.TenantID == "" {
		return msg.Name
	}
	if msg.Name == "" {
		return msg.TenantID
	}
	return fmt.Sprintf("%s:%s:%s", msg.TenantID, msg.InfraType, msg.Name)
}
