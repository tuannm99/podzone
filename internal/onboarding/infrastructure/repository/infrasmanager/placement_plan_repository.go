package repository

import (
	"context"
	"time"

	"github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (s *MongoStore) GetPlacementPlanByRequestID(
	ctx context.Context,
	requestID string,
) (*entity.PlacementPlan, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if requestID == "" {
		return nil, nil
	}

	var doc placementPlanDoc
	err := s.planCol.FindOne(ctx, bson.M{"request_id": requestID}).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	out := doc.toEntity()
	return &out, nil
}

func (s *MongoStore) SavePlacementPlan(ctx context.Context, plan entity.PlacementPlan) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if plan.RequestID == "" {
		return entity.ErrInvalidInput
	}
	now := time.Now().UTC()
	if plan.CreatedAt.IsZero() {
		plan.CreatedAt = now
	}
	plan.UpdatedAt = now

	doc := placementPlanFromEntity(plan)
	_, err := s.planCol.UpdateOne(
		ctx,
		bson.M{"request_id": plan.RequestID},
		bson.M{
			"$set": doc,
			"$setOnInsert": bson.M{
				"created_at": plan.CreatedAt,
			},
		},
		options.Update().SetUpsert(true),
	)
	return err
}

type placementPlanDoc struct {
	RequestID         string                         `bson:"request_id"`
	TenantID          string                         `bson:"tenant_id"`
	StoreID           string                         `bson:"store_id"`
	Runtime           string                         `bson:"runtime"`
	ClusterName       string                         `bson:"cluster_name"`
	Mode              string                         `bson:"mode"`
	DBName            string                         `bson:"db_name"`
	SchemaName        string                         `bson:"schema_name"`
	ProviderMeta      map[string]string              `bson:"provider_meta"`
	InventorySnapshot entity.ResourceInventory       `bson:"inventory_snapshot"`
	CapacitySnapshot  entity.CapacitySnapshot        `bson:"capacity_snapshot"`
	PolicyDecision    entity.PlacementPolicyDecision `bson:"policy_decision"`
	CreatedAt         time.Time                      `bson:"created_at"`
	UpdatedAt         time.Time                      `bson:"updated_at"`
}

func placementPlanFromEntity(plan entity.PlacementPlan) placementPlanDoc {
	return placementPlanDoc{
		RequestID:         plan.RequestID,
		TenantID:          plan.TenantID,
		StoreID:           plan.StoreID,
		Runtime:           string(plan.Runtime),
		ClusterName:       plan.ClusterName,
		Mode:              plan.Mode,
		DBName:            plan.DBName,
		SchemaName:        plan.SchemaName,
		ProviderMeta:      plan.ProviderMeta,
		InventorySnapshot: plan.InventorySnapshot,
		CapacitySnapshot:  plan.CapacitySnapshot,
		PolicyDecision:    plan.PolicyDecision,
		CreatedAt:         plan.CreatedAt,
		UpdatedAt:         plan.UpdatedAt,
	}
}

func (d placementPlanDoc) toEntity() entity.PlacementPlan {
	return entity.PlacementPlan{
		RequestID:         d.RequestID,
		TenantID:          d.TenantID,
		StoreID:           d.StoreID,
		Runtime:           entity.PlacementRuntime(d.Runtime),
		ClusterName:       d.ClusterName,
		Mode:              d.Mode,
		DBName:            d.DBName,
		SchemaName:        d.SchemaName,
		ProviderMeta:      d.ProviderMeta,
		InventorySnapshot: d.InventorySnapshot,
		CapacitySnapshot:  d.CapacitySnapshot,
		PolicyDecision:    d.PolicyDecision,
		CreatedAt:         d.CreatedAt,
		UpdatedAt:         d.UpdatedAt,
	}
}
