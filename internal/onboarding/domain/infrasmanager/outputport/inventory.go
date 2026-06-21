package outputport

import (
	"context"

	"github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/entity"
)

type ResourceInventoryRepository interface {
	LoadResourceInventory(
		ctx context.Context,
		request entity.StorePlacementRequest,
	) (entity.ResourceInventory, error)
}

type CapacityChecker interface {
	CheckPlacementCapacity(
		ctx context.Context,
		request entity.StorePlacementRequest,
		inventory entity.ResourceInventory,
	) (entity.CapacitySnapshot, error)
}

type PlacementPolicyEvaluator interface {
	EvaluatePlacementPolicy(
		ctx context.Context,
		request entity.StorePlacementRequest,
		inventory entity.ResourceInventory,
		capacity entity.CapacitySnapshot,
	) (entity.PlacementPolicyDecision, error)
}
