package resolver

import (
	"context"
	"fmt"
	"strings"

	"github.com/tuannm99/podzone/internal/backoffice/controller/graphql/generated/model"
	routingentity "github.com/tuannm99/podzone/internal/backoffice/domain/routing/entity"
	"github.com/tuannm99/podzone/internal/backoffice/runtime/scope"
)

func toGraphQLRoutedOrder(order routingentity.RoutedOrder) *model.RoutedOrder {
	activities := make([]*model.RoutedOrderActivity, 0, len(order.ActivityLog))
	for _, activity := range order.ActivityLog {
		details := make([]*model.RoutedOrderActivityDetail, 0, len(activity.Details))
		for _, detail := range activity.Details {
			details = append(details, &model.RoutedOrderActivityDetail{
				Key:   detail.Key,
				Value: detail.Value,
			})
		}
		activities = append(activities, &model.RoutedOrderActivity{
			Type:      activity.Type,
			Actor:     activity.Actor,
			Message:   activity.Message,
			Details:   details,
			CreatedAt: activity.CreatedAt,
		})
	}
	return &model.RoutedOrder{
		ID:                     order.ID,
		CandidateID:            order.CandidateID,
		ProductTitle:           order.ProductTitle,
		Partner:                order.Partner,
		Quantity:               order.Quantity,
		Total:                  order.Total,
		CustomerName:           order.CustomerName,
		Status:                 order.Status,
		Timeline:               order.Timeline,
		ActivityLog:            activities,
		ExceptionType:          order.ExceptionType,
		ExceptionStatus:        order.ExceptionStatus,
		ShipmentStatus:         order.ShipmentStatus,
		ShipmentCarrier:        order.ShipmentCarrier,
		ShipmentTrackingNumber: order.ShipmentTrackingNumber,
		ShipmentTrackingURL:    order.ShipmentTrackingURL,
		ShipmentNotes:          order.ShipmentNotes,
		OperatorAssignee:       order.OperatorAssignee,
		ShipmentSLADueAt:       order.ShipmentSlaDueAt,
		IssueSLADueAt:          order.IssueSlaDueAt,
		RoutingBlockCode:       order.RoutingBlockCode,
		RoutingBlockReason:     order.RoutingBlockReason,
		BaseCostSnapshot:       order.BaseCostSnapshot,
		FulfillmentCost:        order.FulfillmentCost,
		ShippingCost:           order.ShippingCost,
		IssueCost:              order.IssueCost,
		IssueResolution:        order.IssueResolution,
		IssueNotes:             order.IssueNotes,
		RealizedMargin:         order.RealizedMargin,
		SettlementStatus:       order.SettlementStatus,
		SettlementNotes:        order.SettlementNotes,
		ShippedAt:              order.ShippedAt,
		DeliveredAt:            order.DeliveredAt,
		CreatedAt:              order.CreatedAt,
		UpdatedAt:              order.UpdatedAt,
	}
}

func toGraphQLRoutedOrderActivity(activity routingentity.RoutedOrderActivity) *model.RoutedOrderActivity {
	details := make([]*model.RoutedOrderActivityDetail, 0, len(activity.Details))
	for _, detail := range activity.Details {
		details = append(details, &model.RoutedOrderActivityDetail{
			Key:   detail.Key,
			Value: detail.Value,
		})
	}
	return &model.RoutedOrderActivity{
		Type:      activity.Type,
		Actor:     activity.Actor,
		Message:   activity.Message,
		Details:   details,
		CreatedAt: activity.CreatedAt,
	}
}

func toGraphQLRoutedOrderActivityFeedEntry(
	entry routingentity.RoutedOrderActivityFeedEntry,
) *model.RoutedOrderActivityFeedEntry {
	return &model.RoutedOrderActivityFeedEntry{
		OrderID:          entry.OrderID,
		ProductTitle:     entry.ProductTitle,
		Partner:          entry.Partner,
		OperatorAssignee: entry.OperatorAssignee,
		Activity:         toGraphQLRoutedOrderActivity(entry.Activity),
	}
}

func toGraphQLRoutedOrderActivityFeedPage(
	page routingentity.RoutedOrderActivityFeedPage,
) *model.RoutedOrderActivityFeedPage {
	entries := make([]*model.RoutedOrderActivityFeedEntry, 0, len(page.Entries))
	for _, entry := range page.Entries {
		entries = append(entries, toGraphQLRoutedOrderActivityFeedEntry(entry))
	}
	return &model.RoutedOrderActivityFeedPage{
		Entries:    entries,
		Total:      page.Total,
		NextCursor: page.NextCursor,
	}
}

func toGraphQLPartnerRoutingProfile(profile routingentity.PartnerRoutingProfile) *model.PartnerRoutingProfile {
	shippingCostRules := make([]*model.PartnerShippingCostRule, 0, len(profile.ShippingCostRules))
	for _, rule := range profile.ShippingCostRules {
		shippingCostRules = append(shippingCostRules, &model.PartnerShippingCostRule{
			Region: rule.Region,
			Cost:   rule.Cost,
		})
	}
	return &model.PartnerRoutingProfile{
		ID:                    profile.ID,
		Code:                  profile.Code,
		Name:                  profile.Name,
		PartnerType:           profile.PartnerType,
		Status:                profile.Status,
		SupportedProductTypes: append([]string(nil), profile.SupportedProductTypes...),
		SupportedRegions:      append([]string(nil), profile.SupportedRegions...),
		SLADays:               int(profile.SLADays),
		RoutingPriority:       int(profile.RoutingPriority),
		BaseFulfillmentCost:   profile.BaseFulfillmentCost,
		ShippingCostRules:     shippingCostRules,
	}
}

func toGraphQLRoutingPartnerOption(option routingentity.RoutingPartnerOption) *model.RoutingPartnerOption {
	return &model.RoutingPartnerOption{
		Partner:                  toGraphQLPartnerRoutingProfile(option.Partner),
		Eligible:                 option.Eligible,
		Reason:                   option.Reason,
		EstimatedFulfillmentCost: option.EstimatedFulfillmentCost,
		EstimatedShippingCost:    option.EstimatedShippingCost,
		EstimatedUnitMargin:      option.EstimatedUnitMargin,
	}
}

func toGraphQLRoutedOrderRecommendation(
	recommendation routingentity.RoutedOrderRecommendation,
) *model.RoutedOrderRecommendation {
	options := make([]*model.RoutingPartnerOption, 0, len(recommendation.Options))
	for _, option := range recommendation.Options {
		options = append(options, toGraphQLRoutingPartnerOption(option))
	}
	return &model.RoutedOrderRecommendation{
		CandidateID:       recommendation.CandidateID,
		ProductTitle:      recommendation.ProductTitle,
		CandidatePartner:  recommendation.CandidatePartner,
		ProductType:       recommendation.ProductType,
		ShipRegion:        recommendation.ShipRegion,
		SelectedPartner:   recommendation.SelectedPartner,
		BlockedReasonCode: recommendation.BlockedReasonCode,
		BlockedReason:     recommendation.BlockedReason,
		Summary:           recommendation.Summary,
		Options:           options,
	}
}

func stringOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func boolOrFalse(value *bool) bool {
	if value == nil {
		return false
	}
	return *value
}

func requiredStoreID(ctx context.Context) (string, error) {
	storeID := strings.TrimSpace(scope.CurrentStoreID(ctx))
	if storeID == "" {
		return "", fmt.Errorf("store scope is required")
	}
	return storeID, nil
}
