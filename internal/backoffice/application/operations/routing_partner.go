package operations

import (
	"context"
	"fmt"
	"strings"

	catalogctx "github.com/tuannm99/podzone/internal/backoffice/domain/catalog"
	orderctx "github.com/tuannm99/podzone/internal/backoffice/domain/order"
	routingctx "github.com/tuannm99/podzone/internal/backoffice/domain/routing"
	"github.com/tuannm99/podzone/pkg/ddd"
	"github.com/tuannm99/podzone/pkg/toolkit"
)

type RoutingInteractor struct {
	orders         routingctx.OrderRoutingRepository
	customerOrders orderctx.CustomerOrderQueryRepository
	products       catalogctx.ProductSetupRepository
	partners       routingctx.PartnerDirectory
	events         ddd.EventDispatcher
	clock          ddd.Clock
}

var (
	_ routingctx.RoutingCommandUsecase = (*RoutingInteractor)(nil)
	_ routingctx.RoutingQueryUsecase   = (*RoutingInteractor)(nil)
)

func NewRoutingInteractor(
	orders routingctx.OrderRoutingRepository,
	customerOrders orderctx.CustomerOrderQueryRepository,
	products catalogctx.ProductSetupRepository,
	partners routingctx.PartnerDirectory,
	dispatcher ddd.EventDispatcher,
	clock ddd.Clock,
) *RoutingInteractor {
	return &RoutingInteractor{
		orders:         orders,
		customerOrders: customerOrders,
		products:       products,
		partners:       partners,
		events:         dispatcher,
		clock:          clock,
	}
}

func (i *RoutingInteractor) RecommendRoutedOrderPartner(
	ctx context.Context,
	query routingctx.RecommendRoutedOrderPartnerQuery,
) (*routingctx.RoutedOrderRecommendation, error) {
	storeID, err := routingctx.RequiredStoreScope(ctx, query.StoreID)
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
	return routingctx.BuildRoutingRecommendation(
		candidate,
		partners,
		routingctx.NormalizeRoutingLabel(query.ProductType),
		routingctx.NormalizeRoutingLabel(query.ShipRegion),
		strings.TrimSpace(query.PreferredPartner),
		i.clock.Now(),
	), nil
}

func (i *RoutingInteractor) ForceRerouteBlockedOrder(
	ctx context.Context,
	cmd routingctx.ForceRerouteBlockedOrderCmd,
) (*routingctx.RoutedOrder, error) {
	storeID, err := routingctx.RequiredStoreScope(ctx, cmd.StoreID)
	if err != nil {
		return nil, err
	}
	order, err := i.orders.GetByID(ctx, strings.TrimSpace(cmd.OrderID))
	if err != nil {
		return nil, err
	}
	if err := routingctx.EnsureOrderStore(order, storeID); err != nil {
		return nil, err
	}
	customerOrder, err := i.customerOrders.GetCustomerOrder(ctx, storeID, order.ID)
	if err != nil {
		return nil, err
	}
	if order.Status != routingctx.RoutedOrderStatusRoutingBlocked {
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
	recommendation := routingctx.BuildRoutingRecommendation(
		candidate,
		partners,
		routingctx.NormalizeRoutingLabel(routingctx.OrderRoutingLabel(order, candidate)),
		routingctx.NormalizeRoutingLabel(routingctx.ShipRegionFromOrder(order)),
		preferredPartner,
		i.clock.Now(),
	)
	if recommendation.SelectedPartner == "" {
		return nil, fmt.Errorf("preferred partner %s is not eligible for reroute", preferredPartner)
	}
	selectedOption := routingctx.FindSelectedRoutingOption(recommendation)
	if selectedOption == nil {
		return nil, fmt.Errorf("selected routing option not found")
	}

	now := i.clock.Now()
	domainEvents, err := applyManualReroute(
		order,
		customerOrder,
		recommendation.SelectedPartner,
		selectedOption.EstimatedFulfillmentCost,
		selectedOption.EstimatedShippingCost,
		selectedOption.EstimatedUnitMargin,
		recommendation.Summary,
		routingctx.ActivityActorFromContext(ctx),
		now,
	)
	if err != nil {
		return nil, err
	}
	saved, err := i.orders.Update(ctx, *order)
	if err != nil {
		return nil, err
	}
	if err := dispatchDomainEvents(ctx, i.events, domainEvents); err != nil {
		return nil, err
	}
	return saved, nil
}
