package outputport

import (
	"context"

	"github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/entity"
)

type PlacementPlanner interface {
	PlanStorePlacement(
		ctx context.Context,
		request entity.StorePlacementRequest,
	) (entity.PlacementPlan, error)
}

type StorageProvisioner interface {
	ProvisionStorePlacement(
		ctx context.Context,
		request entity.StorePlacementRequest,
		plan entity.PlacementPlan,
	) (entity.PlacementAllocation, error)
}

type PlacementRepository interface {
	GetPlacementAllocation(
		ctx context.Context,
		tenantID string,
		storeID string,
	) (*entity.PlacementAllocation, error)
	SavePlacementAllocation(ctx context.Context, allocation entity.PlacementAllocation) error
}

type PlacementPlanRepository interface {
	GetPlacementPlanByRequestID(ctx context.Context, requestID string) (*entity.PlacementPlan, error)
	SavePlacementPlan(ctx context.Context, plan entity.PlacementPlan) error
}

type PlacementRouteReader interface {
	IsPlacementRouteReady(ctx context.Context, tenantID string) (bool, error)
}

type PlacementRouteWriter interface {
	PublishPlacementRoute(
		ctx context.Context,
		tenantID string,
		allocation entity.PlacementAllocation,
	) error
}
