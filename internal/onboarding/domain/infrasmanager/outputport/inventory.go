package outputport

import (
	"context"

	"github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/entity"
	"github.com/tuannm99/podzone/pkg/collection"
)

type ResourceInventoryRepository interface {
	LoadResourceInventory(
		ctx context.Context,
		request entity.StorePlacementRequest,
	) (entity.ResourceInventory, error)
	ListDatabaseClusters(
		ctx context.Context,
		query collection.Query,
	) (collection.Page[entity.DatabaseCluster], error)
	GetDatabaseCluster(ctx context.Context, name string) (*entity.DatabaseCluster, error)
	UpsertDatabaseCluster(ctx context.Context, cluster entity.DatabaseCluster) error
	UpdateDatabaseClusterHealth(
		ctx context.Context,
		name string,
		health entity.DatabaseClusterHealth,
	) error
	DeleteDatabaseCluster(ctx context.Context, name string) error
	ListKubernetesClusters(
		ctx context.Context,
		query collection.Query,
	) (collection.Page[entity.KubernetesCluster], error)
	UpsertKubernetesCluster(ctx context.Context, cluster entity.KubernetesCluster) error
	DeleteKubernetesCluster(ctx context.Context, name string) error
	ListRuntimePools(
		ctx context.Context,
		query collection.Query,
	) (collection.Page[entity.RuntimePool], error)
	UpsertRuntimePool(ctx context.Context, pool entity.RuntimePool) error
	DeleteRuntimePool(ctx context.Context, name string) error
}

type CapacityChecker interface {
	CheckPlacementCapacity(
		ctx context.Context,
		request entity.StorePlacementRequest,
		inventory entity.ResourceInventory,
	) (entity.CapacitySnapshot, error)
}

type ResourceHealthChecker interface {
	CheckDatabaseClusterHealth(
		ctx context.Context,
		cluster entity.DatabaseCluster,
	) (entity.DatabaseClusterHealth, error)
}

type PlacementPolicyEvaluator interface {
	EvaluatePlacementPolicy(
		ctx context.Context,
		request entity.StorePlacementRequest,
		inventory entity.ResourceInventory,
		capacity entity.CapacitySnapshot,
	) (entity.PlacementPolicyDecision, error)
}
