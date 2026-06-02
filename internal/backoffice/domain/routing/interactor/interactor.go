package interactor

import (
	"context"
	"fmt"
	"strings"
	"time"

	catalogoutputport "github.com/tuannm99/podzone/internal/backoffice/domain/catalog/outputport"
	routingentity "github.com/tuannm99/podzone/internal/backoffice/domain/routing/entity"
	routinginputport "github.com/tuannm99/podzone/internal/backoffice/domain/routing/inputport"
	routingoutputport "github.com/tuannm99/podzone/internal/backoffice/domain/routing/outputport"
	routingpolicy "github.com/tuannm99/podzone/internal/backoffice/domain/routing/policy"
	routingsupport "github.com/tuannm99/podzone/internal/backoffice/domain/routing/support"
	"github.com/tuannm99/podzone/pkg/toolkit"
)

type Interactor struct {
	orders   routingoutputport.OrderRoutingRepository
	products catalogoutputport.ProductSetupRepository
	partners routingoutputport.PartnerDirectory
}

var (
	_ routinginputport.RoutingCommandUsecase = (*Interactor)(nil)
	_ routinginputport.RoutingQueryUsecase   = (*Interactor)(nil)
)

func New(
	orders routingoutputport.OrderRoutingRepository,
	products catalogoutputport.ProductSetupRepository,
	partners routingoutputport.PartnerDirectory,
) *Interactor {
	return &Interactor{orders: orders, products: products, partners: partners}
}

func (i *Interactor) RecommendRoutedOrderPartner(
	ctx context.Context,
	query routinginputport.RecommendRoutedOrderPartnerQuery,
) (*routingentity.RoutedOrderRecommendation, error) {
	storeID, err := routingsupport.RequiredStoreScope(ctx, query.StoreID)
	if err != nil {
		return nil, err
	}
	candidateID := strings.TrimSpace(query.CandidateID)
	if candidateID == "" {
		return nil, fmt.Errorf("candidate id is required")
	}
	candidate, err := i.products.GetCandidateByID(ctx, storeID, candidateID)
	if err != nil {
		return nil, err
	}
	if candidate == nil {
		return nil, fmt.Errorf("product candidate not found")
	}
	tenantID, err := toolkit.GetTenantID(ctx)
	if err != nil {
		return nil, err
	}
	partners, err := i.partners.ListActivePartners(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	return routingpolicy.BuildRoutingRecommendation(
		candidate,
		partners,
		routingpolicy.NormalizeRoutingLabel(query.ProductType),
		routingpolicy.NormalizeRoutingLabel(query.ShipRegion),
		strings.TrimSpace(query.PreferredPartner),
	), nil
}

func (i *Interactor) ForceRerouteBlockedOrder(
	ctx context.Context,
	cmd routinginputport.ForceRerouteBlockedOrderCmd,
) (*routingentity.RoutedOrder, error) {
	storeID, err := routingsupport.RequiredStoreScope(ctx, cmd.StoreID)
	if err != nil {
		return nil, err
	}
	order, err := i.orders.GetByID(ctx, strings.TrimSpace(cmd.OrderID))
	if err != nil {
		return nil, err
	}
	if err := routingsupport.EnsureOrderStore(order, storeID); err != nil {
		return nil, err
	}
	if order.Status != routingentity.RoutedOrderStatusRoutingBlocked {
		return nil, fmt.Errorf("routed order is not in routing_blocked status")
	}
	preferredPartner := strings.TrimSpace(cmd.PreferredPartner)
	if preferredPartner == "" {
		return nil, fmt.Errorf("preferred partner is required")
	}

	candidate, err := i.products.GetCandidateByID(ctx, storeID, strings.TrimSpace(order.CandidateID))
	if err != nil {
		return nil, err
	}
	if candidate == nil {
		return nil, fmt.Errorf("product candidate not found")
	}
	tenantID, err := toolkit.GetTenantID(ctx)
	if err != nil {
		return nil, err
	}
	partners, err := i.partners.ListActivePartners(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	recommendation := routingpolicy.BuildRoutingRecommendation(
		candidate,
		partners,
		routingpolicy.NormalizeRoutingLabel(routingsupport.OrderRoutingLabel(order, candidate)),
		routingpolicy.NormalizeRoutingLabel(routingsupport.ShipRegionFromOrder(order)),
		preferredPartner,
	)
	if recommendation.SelectedPartner == "" {
		return nil, fmt.Errorf("preferred partner %s is not eligible for reroute", preferredPartner)
	}
	selectedOption := routingpolicy.FindSelectedRoutingOption(recommendation)
	if selectedOption == nil {
		return nil, fmt.Errorf("selected routing option not found")
	}

	now := time.Now().UTC()
	if err := order.ApplyManualReroute(
		recommendation.SelectedPartner,
		selectedOption.EstimatedFulfillmentCost,
		selectedOption.EstimatedShippingCost,
		selectedOption.EstimatedUnitMargin,
		recommendation.Summary,
		routingsupport.ActivityActorFromContext(ctx),
		now,
	); err != nil {
		return nil, err
	}
	return i.orders.Update(ctx, *order)
}
