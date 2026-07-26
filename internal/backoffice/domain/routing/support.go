package routing

import (
	"context"
	"strings"

	catalogentity "github.com/tuannm99/podzone/internal/backoffice/domain/catalog"
	"github.com/tuannm99/podzone/internal/backoffice/runtime/scope"
	"github.com/tuannm99/podzone/pkg/toolkit"
)

// RequiredStoreScope resolves the store ID a caller may act on. explicitStoreID
// must agree with the request scope IAM already checked permission against
// (scope.CurrentStoreID) - it is not a second source of truth. See MT5,
// docs/07-problems/multitenancy-2026-07-23.md: an explicit value that wins
// unconditionally would let a future DTO check permission on store A and
// act on store B.
func RequiredStoreScope(ctx context.Context, explicitStoreID string) (string, error) {
	explicit := strings.TrimSpace(explicitStoreID)
	scoped := strings.TrimSpace(scope.CurrentStoreID(ctx))
	switch {
	case explicit == "":
		explicit = scoped
	case scoped != "" && explicit != scoped:
		return "", ErrStoreScopeMismatch
	}
	if explicit == "" {
		return "", ErrStoreScopeRequired
	}
	return explicit, nil
}

func EnsureOrderStore(order *RoutedOrder, storeID string) error {
	if order == nil {
		return ErrRoutedOrderNotFound
	}
	if strings.TrimSpace(order.StoreID) == "" || strings.TrimSpace(order.StoreID) != strings.TrimSpace(storeID) {
		return ErrRoutedOrderNotFound
	}
	return nil
}

func ActivityActorFromContext(ctx context.Context) string {
	userID, err := toolkit.GetUserID(ctx)
	if err != nil {
		return "system"
	}
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return "system"
	}
	return "user:" + userID
}

func OrderRoutingLabel(order *RoutedOrder, candidate *catalogentity.ProductSetupCandidate) string {
	for _, activity := range order.ActivityLog {
		for _, detail := range activity.Details {
			if detail.Key == "product_type" && strings.TrimSpace(detail.Value) != "" {
				return detail.Value
			}
		}
	}
	if candidate != nil {
		return InferProductType(candidate.Title)
	}
	return ""
}

func ShipRegionFromOrder(order *RoutedOrder) string {
	for _, activity := range order.ActivityLog {
		for _, detail := range activity.Details {
			if detail.Key == "ship_region" && strings.TrimSpace(detail.Value) != "" {
				return detail.Value
			}
		}
	}
	return ""
}

func InferProductType(title string) string {
	normalized := strings.ToLower(strings.TrimSpace(title))
	switch {
	case strings.Contains(normalized, "hoodie"):
		return "hoodie"
	case strings.Contains(normalized, "poster"):
		return "poster"
	case strings.Contains(normalized, "tote"):
		return "tote"
	case strings.Contains(normalized, "tee"), strings.Contains(normalized, "shirt"):
		return "tshirt"
	default:
		return ""
	}
}
