package infrasmanager

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/entity"
	"github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/inputport"
)

func (s *Interactor) GetTenantPlacementStatus(
	ctx context.Context,
	tenantID string,
) (*inputport.PlacementStatus, error) {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return nil, entity.ErrInvalidInput
	}
	if s.placements == nil {
		return nil, errors.New("placement repository is not configured")
	}

	allocation, err := s.placements.GetTenantPlacementAllocation(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	if allocation == nil {
		return nil, entity.ErrPlacementNotFound
	}

	var route *entity.PlacementRoute
	if s.routeReader != nil {
		route, err = s.routeReader.GetPlacementRoute(ctx, tenantID)
		if err != nil {
			return nil, err
		}
	}
	status := placementStatusFromAllocation(tenantID, *allocation, route)
	return &status, nil
}

func (s *Interactor) ReconcileTenantPlacement(
	ctx context.Context,
	tenantID string,
	actor map[string]string,
) (*inputport.PlacementReconcileResponse, error) {
	_ = actor
	status, err := s.GetTenantPlacementStatus(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	if !status.NeedsRepair {
		return &inputport.PlacementReconcileResponse{
			Status:     *status,
			KVStoreKey: placementRouteKey(tenantID),
		}, nil
	}
	if s.placements == nil || s.routeWriter == nil {
		return nil, errors.New("placement route writer is not configured")
	}
	allocation, err := s.placements.GetTenantPlacementAllocation(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	if allocation == nil {
		return nil, entity.ErrPlacementNotFound
	}
	if err := s.routeWriter.PublishPlacementRoute(ctx, tenantID, *allocation); err != nil {
		return nil, err
	}

	publishedAt := time.Now().UTC()
	reconciled := placementStatusFromAllocation(
		strings.TrimSpace(tenantID),
		*allocation,
		allocationPlacementRoute(*allocation),
	)
	return &inputport.PlacementReconcileResponse{
		Status:      reconciled,
		Repaired:    true,
		KVStoreKey:  placementRouteKey(tenantID),
		PublishedAt: &publishedAt,
	}, nil
}

func placementStatusFromAllocation(
	tenantID string,
	allocation entity.PlacementAllocation,
	route *entity.PlacementRoute,
) inputport.PlacementStatus {
	allocationRoute := allocationPlacementRoute(allocation)
	routeReady := route != nil
	inSync := routeReady && placementRoutesEqual(*allocationRoute, *route)
	reason := ""
	if !routeReady {
		reason = "placement route is missing"
	} else if !inSync {
		reason = "placement route differs from allocation"
	}

	return inputport.PlacementStatus{
		TenantID:        tenantID,
		AllocationID:    allocation.ID,
		AllocationReady: strings.EqualFold(allocation.Status, "ready"),
		RouteReady:      routeReady,
		InSync:          inSync,
		NeedsRepair:     !inSync,
		Reason:          reason,
		Allocation:      toInputPlacementRoute(allocationRoute),
		Route:           toInputPlacementRoute(route),
		UpdatedAt:       allocation.UpdatedAt,
	}
}

func allocationPlacementRoute(allocation entity.PlacementAllocation) *entity.PlacementRoute {
	return &entity.PlacementRoute{
		ClusterName: allocation.ClusterName,
		Mode:        allocation.Mode,
		DBName:      allocation.DBName,
		SchemaName:  allocation.SchemaName,
	}
}

func placementRoutesEqual(left entity.PlacementRoute, right entity.PlacementRoute) bool {
	return left.ClusterName == right.ClusterName &&
		left.Mode == right.Mode &&
		left.DBName == right.DBName &&
		left.SchemaName == right.SchemaName
}

func toInputPlacementRoute(route *entity.PlacementRoute) *inputport.PlacementRoute {
	if route == nil {
		return nil
	}
	return &inputport.PlacementRoute{
		ClusterName: route.ClusterName,
		Mode:        route.Mode,
		DBName:      route.DBName,
		SchemaName:  route.SchemaName,
	}
}

func placementRouteKey(tenantID string) string {
	return "podzone/tenants/" + strings.TrimSpace(tenantID) + "/placement"
}
