package routing

import (
	"fmt"
	"sort"
	"strings"
	"time"

	catalogentity "github.com/tuannm99/podzone/internal/backoffice/domain/catalog/entity"
	routingentity "github.com/tuannm99/podzone/internal/backoffice/domain/routing/entity"
	routinginputport "github.com/tuannm99/podzone/internal/backoffice/domain/routing/inputport"
)

func newActivity(
	activityType,
	actor,
	message string,
	createdAt time.Time,
	details []routingentity.RoutedOrderActivityDetail,
) routingentity.RoutedOrderActivity {
	return routingentity.RoutedOrderActivity{
		Type:      strings.TrimSpace(activityType),
		Actor:     strings.TrimSpace(actor),
		Message:   strings.TrimSpace(message),
		Details:   details,
		CreatedAt: createdAt,
	}
}

func activityDetails(pairs ...string) []routingentity.RoutedOrderActivityDetail {
	details := make([]routingentity.RoutedOrderActivityDetail, 0, len(pairs)/2)
	for i := 0; i+1 < len(pairs); i += 2 {
		key := strings.TrimSpace(pairs[i])
		value := strings.TrimSpace(pairs[i+1])
		if key == "" || value == "" {
			continue
		}
		details = append(details, routingentity.RoutedOrderActivityDetail{Key: key, Value: value})
	}
	return details
}

func formatOptionalTime(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}

func routingTimelineEntry(status, partner string) string {
	switch status {
	case routingentity.RoutedOrderStatusInProduction:
		return fmt.Sprintf("Sent to %s for POD production", partner)
	case routingentity.RoutedOrderStatusShipped:
		return fmt.Sprintf("Marked as shipped from %s", partner)
	default:
		return fmt.Sprintf("Queued for %s", partner)
	}
}

func normalizeExceptionType(raw string) string {
	switch strings.TrimSpace(raw) {
	case "artwork_issue", "partner_delay", "address_hold", "reprint_request":
		return raw
	default:
		return ""
	}
}

func normalizeExceptionStatus(raw string) string {
	switch strings.TrimSpace(raw) {
	case routingentity.RoutedOrderExceptionStatusOpen,
		routingentity.RoutedOrderExceptionStatusEscalated,
		routingentity.RoutedOrderExceptionStatusResolved:
		return raw
	default:
		return ""
	}
}

func normalizeShipmentStatus(raw string) string {
	switch strings.TrimSpace(raw) {
	case routingentity.RoutedOrderShipmentStatusAwaitingLabel,
		routingentity.RoutedOrderShipmentStatusLabelReady,
		routingentity.RoutedOrderShipmentStatusInTransit,
		routingentity.RoutedOrderShipmentStatusDelivered,
		routingentity.RoutedOrderShipmentStatusDeliveryIssue:
		return raw
	default:
		return ""
	}
}

func normalizeSettlementStatus(raw string) string {
	switch strings.TrimSpace(raw) {
	case routingentity.RoutedOrderSettlementStatusPending,
		routingentity.RoutedOrderSettlementStatusReconciled,
		routingentity.RoutedOrderSettlementStatusPaid,
		routingentity.RoutedOrderSettlementStatusDisputed:
		return raw
	default:
		return ""
	}
}

func normalizeIssueResolution(raw string) string {
	switch strings.TrimSpace(raw) {
	case routingentity.RoutedOrderIssueResolutionMonitor,
		routingentity.RoutedOrderIssueResolutionReprint,
		routingentity.RoutedOrderIssueResolutionRefund,
		routingentity.RoutedOrderIssueResolutionCarrierClaim,
		routingentity.RoutedOrderIssueResolutionAddressRetry:
		return raw
	default:
		return ""
	}
}

func shipmentTimelineEntry(order *routingentity.RoutedOrder) string {
	switch order.ShipmentStatus {
	case routingentity.RoutedOrderShipmentStatusLabelReady:
		return fmt.Sprintf("Shipment prepared with %s", fallbackShipmentCarrier(order.ShipmentCarrier))
	case routingentity.RoutedOrderShipmentStatusInTransit:
		return fmt.Sprintf(
			"Shipment is in transit via %s%s",
			fallbackShipmentCarrier(order.ShipmentCarrier),
			fallbackTrackingSuffix(order.ShipmentTrackingNumber),
		)
	case routingentity.RoutedOrderShipmentStatusDelivered:
		return "Shipment marked delivered by store operator"
	case routingentity.RoutedOrderShipmentStatusDeliveryIssue:
		return "Shipment issue flagged for manual follow-up"
	default:
		return "Shipment is awaiting label assignment"
	}
}

func settlementTimelineEntry(order *routingentity.RoutedOrder) string {
	switch order.SettlementStatus {
	case routingentity.RoutedOrderSettlementStatusPaid:
		return fmt.Sprintf("Settlement marked paid with realized margin %s", order.RealizedMargin)
	case routingentity.RoutedOrderSettlementStatusDisputed:
		return "Settlement flagged for manual dispute follow-up"
	case routingentity.RoutedOrderSettlementStatusReconciled:
		return fmt.Sprintf("Settlement reconciled with realized margin %s", order.RealizedMargin)
	default:
		return fmt.Sprintf("Settlement remains pending with current margin %s", order.RealizedMargin)
	}
}

func issueHandlingTimelineEntry(order *routingentity.RoutedOrder) string {
	return fmt.Sprintf(
		"Issue handling updated: %s with impact %s",
		strings.ReplaceAll(order.IssueResolution, "_", " "),
		order.IssueCost,
	)
}

func queueControlTimelineEntry(order *routingentity.RoutedOrder) string {
	shipmentDue := "none"
	if order.ShipmentSlaDueAt != nil {
		shipmentDue = order.ShipmentSlaDueAt.UTC().Format(time.RFC3339)
	}
	issueDue := "none"
	if order.IssueSlaDueAt != nil {
		issueDue = order.IssueSlaDueAt.UTC().Format(time.RFC3339)
	}
	return fmt.Sprintf(
		"Queue ownership updated: %s · shipment SLA %s · issue SLA %s",
		order.OperatorAssignee,
		shipmentDue,
		issueDue,
	)
}

func bulkUpdateTimelineEntry(order *routingentity.RoutedOrder, cmd routinginputport.BulkUpdateRoutedOrdersCmd) string {
	parts := make([]string, 0, 3)
	if cmd.OperatorAssignee != nil {
		parts = append(parts, fmt.Sprintf("owner %s", order.OperatorAssignee))
	}
	if cmd.ShipmentSlaDueAt != nil {
		parts = append(parts, fmt.Sprintf("shipment SLA %s", cmd.ShipmentSlaDueAt.UTC().Format(time.RFC3339)))
	}
	if cmd.SettlementStatus != nil {
		parts = append(parts, fmt.Sprintf("settlement %s", order.SettlementStatus))
	}
	return fmt.Sprintf("Bulk queue update applied: %s", strings.Join(parts, " · "))
}

func fallbackShipmentCarrier(carrier string) string {
	carrier = strings.TrimSpace(carrier)
	if carrier == "" {
		return "manual carrier"
	}
	return carrier
}

func fallbackTrackingSuffix(trackingNumber string) string {
	trackingNumber = strings.TrimSpace(trackingNumber)
	if trackingNumber == "" {
		return ""
	}
	return fmt.Sprintf(" (%s)", trackingNumber)
}

func buildRoutingRecommendation(
	candidate *catalogentity.ProductSetupCandidate,
	partners []routingentity.PartnerRoutingProfile,
	productType string,
	shipRegion string,
	preferredPartner string,
) *routingentity.RoutedOrderRecommendation {
	recommendation := &routingentity.RoutedOrderRecommendation{
		CandidateID:      candidate.ID,
		ProductTitle:     candidate.Title,
		CandidatePartner: strings.TrimSpace(candidate.Partner),
		ProductType:      productType,
		ShipRegion:       shipRegion,
		Options:          make([]routingentity.RoutingPartnerOption, 0, len(partners)),
	}

	for _, partner := range partners {
		eligible, reason := evaluatePartnerEligibility(partner, productType, shipRegion)
		estimatedFulfillmentCost := estimateFulfillmentCost(candidate.BaseCost, partner.BaseFulfillmentCost)
		estimatedShippingCost := estimateShippingCost(partner.ShippingCostRules, shipRegion)
		estimatedUnitMargin := estimateUnitMargin(
			candidate.RetailPrice,
			candidate.BaseCost,
			partner,
			shipRegion,
		)
		recommendation.Options = append(recommendation.Options, routingentity.RoutingPartnerOption{
			Partner:                  partner,
			Eligible:                 eligible,
			Reason:                   describeRoutingOption(reason, estimatedUnitMargin),
			EstimatedFulfillmentCost: estimatedFulfillmentCost,
			EstimatedShippingCost:    estimatedShippingCost,
			EstimatedUnitMargin:      estimatedUnitMargin,
		})
	}

	sort.SliceStable(recommendation.Options, func(i, j int) bool {
		left := recommendation.Options[i]
		right := recommendation.Options[j]
		if left.Eligible != right.Eligible {
			return left.Eligible
		}
		leftRank := routingOptionRank(left)
		rightRank := routingOptionRank(right)
		if leftRank != rightRank {
			return leftRank < rightRank
		}
		leftMargin, leftMarginOK := parseMoney(left.EstimatedUnitMargin)
		rightMargin, rightMarginOK := parseMoney(right.EstimatedUnitMargin)
		if leftMarginOK && rightMarginOK && leftMargin != rightMargin {
			return leftMargin > rightMargin
		}
		if left.Partner.SLADays != right.Partner.SLADays {
			return left.Partner.SLADays < right.Partner.SLADays
		}
		if left.Partner.RoutingPriority != right.Partner.RoutingPriority {
			return left.Partner.RoutingPriority > right.Partner.RoutingPriority
		}
		return strings.ToLower(left.Partner.Name) < strings.ToLower(right.Partner.Name)
	})

	preferredPartner = strings.TrimSpace(preferredPartner)
	if preferredPartner != "" {
		for _, option := range recommendation.Options {
			if !option.Eligible {
				continue
			}
			if strings.EqualFold(option.Partner.Name, preferredPartner) ||
				strings.EqualFold(option.Partner.Code, preferredPartner) {
				recommendation.SelectedPartner = option.Partner.Name
				recommendation.Summary = fmt.Sprintf(
					"Preferred partner %s is eligible for %s in %s with expected unit margin %s.",
					option.Partner.Name,
					fallbackRoutingLabel(productType, "any product"),
					fallbackRoutingLabel(shipRegion, "any region"),
					option.EstimatedUnitMargin,
				)
				return recommendation
			}
		}
		recommendation.Summary = fmt.Sprintf(
			"Preferred partner %s is not eligible for %s in %s.",
			preferredPartner,
			fallbackRoutingLabel(productType, "any product"),
			fallbackRoutingLabel(shipRegion, "any region"),
		)
	}

	for _, option := range recommendation.Options {
		if option.Eligible && routingOptionRank(option) <= routingOptionRankUnknownMargin {
			recommendation.SelectedPartner = option.Partner.Name
			recommendation.Summary = summarizeMarginAwareSelection(recommendation.CandidatePartner, option)
			return recommendation
		}
	}

	if hasEligibleOptions(recommendation.Options) {
		recommendation.BlockedReasonCode = "negative_margin"
		recommendation.BlockedReason = "all eligible partners have negative expected margin"
		recommendation.Summary = fmt.Sprintf(
			"No auto-route partner selected for %s in %s because all eligible options have negative expected margin.",
			fallbackRoutingLabel(productType, "the requested product"),
			fallbackRoutingLabel(shipRegion, "the requested region"),
		)
		return recommendation
	}

	recommendation.BlockedReasonCode, recommendation.BlockedReason = deriveRoutingBlockReason(recommendation.Options)
	recommendation.Summary = fmt.Sprintf(
		"No active partner matches %s in %s.",
		fallbackRoutingLabel(productType, "the requested product"),
		fallbackRoutingLabel(shipRegion, "the requested region"),
	)
	return recommendation
}

func evaluatePartnerEligibility(
	partner routingentity.PartnerRoutingProfile,
	productType string,
	shipRegion string,
) (bool, string) {
	if partner.Status != "active" {
		return false, "partner is inactive"
	}
	if len(partner.SupportedProductTypes) > 0 && productType != "" &&
		!containsNormalized(partner.SupportedProductTypes, productType) {
		return false, fmt.Sprintf("does not support %s", productType)
	}
	if len(partner.SupportedRegions) > 0 && shipRegion != "" &&
		!containsNormalized(partner.SupportedRegions, shipRegion) {
		return false, fmt.Sprintf("does not ship to %s", shipRegion)
	}
	return true, fmt.Sprintf("eligible with priority %d and %d-day SLA", partner.RoutingPriority, partner.SLADays)
}

const (
	routingOptionRankProfitable = iota
	routingOptionRankUnknownMargin
	routingOptionRankNegativeMargin
	routingOptionRankIneligible
)

func routingOptionRank(option routingentity.RoutingPartnerOption) int {
	if !option.Eligible {
		return routingOptionRankIneligible
	}
	margin, ok := parseMoney(option.EstimatedUnitMargin)
	if !ok {
		return routingOptionRankUnknownMargin
	}
	if margin < 0 {
		return routingOptionRankNegativeMargin
	}
	return routingOptionRankProfitable
}

func hasEligibleOptions(options []routingentity.RoutingPartnerOption) bool {
	for _, option := range options {
		if option.Eligible {
			return true
		}
	}
	return false
}

func summarizeMarginAwareSelection(candidatePartner string, option routingentity.RoutingPartnerOption) string {
	switch {
	case candidatePartner != "" && strings.EqualFold(option.Partner.Name, candidatePartner):
		return fmt.Sprintf(
			"Candidate default partner %s remains best with expected unit margin %s, %d-day SLA, and routing priority %d.",
			option.Partner.Name,
			option.EstimatedUnitMargin,
			option.Partner.SLADays,
			option.Partner.RoutingPriority,
		)
	case candidatePartner != "":
		return fmt.Sprintf(
			"Recommended %s over candidate default %s based on expected unit margin %s, %d-day SLA, and routing priority %d.",
			option.Partner.Name,
			candidatePartner,
			option.EstimatedUnitMargin,
			option.Partner.SLADays,
			option.Partner.RoutingPriority,
		)
	case parseableMargin(option.EstimatedUnitMargin):
		return fmt.Sprintf(
			"Recommended %s based on expected unit margin %s, %d-day SLA, and routing priority %d.",
			option.Partner.Name,
			option.EstimatedUnitMargin,
			option.Partner.SLADays,
			option.Partner.RoutingPriority,
		)
	default:
		return fmt.Sprintf(
			"Recommended %s based on capability match, %d-day SLA, and routing priority %d; expected margin is unavailable.",
			option.Partner.Name,
			option.Partner.SLADays,
			option.Partner.RoutingPriority,
		)
	}
}

func describeRoutingOption(baseReason, estimatedUnitMargin string) string {
	if !parseableMargin(estimatedUnitMargin) {
		return fmt.Sprintf("%s; expected unit margin unavailable", baseReason)
	}
	margin, _ := parseMoney(estimatedUnitMargin)
	if margin < 0 {
		return fmt.Sprintf("%s; negative expected unit margin %s", baseReason, estimatedUnitMargin)
	}
	return fmt.Sprintf("%s; expected unit margin %s", baseReason, estimatedUnitMargin)
}

func parseableMargin(raw string) bool {
	_, ok := parseMoney(raw)
	return ok
}

func findSelectedRoutingOption(
	recommendation *routingentity.RoutedOrderRecommendation,
) *routingentity.RoutingPartnerOption {
	if recommendation == nil || strings.TrimSpace(recommendation.SelectedPartner) == "" {
		return nil
	}
	for idx := range recommendation.Options {
		option := &recommendation.Options[idx]
		if strings.EqualFold(option.Partner.Name, recommendation.SelectedPartner) ||
			strings.EqualFold(option.Partner.Code, recommendation.SelectedPartner) {
			return option
		}
	}
	return nil
}

func deriveRoutingBlockReason(options []routingentity.RoutingPartnerOption) (string, string) {
	if len(options) == 0 {
		return "no_partner_profile", "no active partner profile is available for routing"
	}
	for _, option := range options {
		if strings.Contains(option.Reason, "does not support") {
			return "unsupported_product", "no partner supports the requested product type"
		}
	}
	for _, option := range options {
		if strings.Contains(option.Reason, "does not ship to") {
			return "unsupported_region", "no partner serves the requested shipping region"
		}
	}
	for _, option := range options {
		if strings.Contains(option.Reason, "inactive") {
			return "inactive_partner", "matching partners are inactive"
		}
	}
	return "no_eligible_partner", "no eligible partner matched the routing request"
}

func estimateFulfillmentCost(candidateBaseCost, partnerBaseCost string) string {
	if strings.TrimSpace(partnerBaseCost) != "" {
		return partnerBaseCost
	}
	if strings.TrimSpace(candidateBaseCost) != "" {
		return candidateBaseCost
	}
	return "TBD"
}

func estimateShippingCost(rules []routingentity.PartnerShippingCostRule, shipRegion string) string {
	normalizedRegion := normalizeRoutingLabel(shipRegion)
	for _, rule := range rules {
		if normalizeRoutingLabel(rule.Region) == normalizedRegion && strings.TrimSpace(rule.Cost) != "" {
			return rule.Cost
		}
	}
	return "$0.00"
}

func estimateUnitMargin(
	retailPrice string,
	candidateBaseCost string,
	partner routingentity.PartnerRoutingProfile,
	shipRegion string,
) string {
	fulfillmentCost := estimateFulfillmentCost(candidateBaseCost, partner.BaseFulfillmentCost)
	shippingCost := estimateShippingCost(partner.ShippingCostRules, shipRegion)
	return calculateMargin(retailPrice, fulfillmentCost, shippingCost, "$0.00")
}

func containsNormalized(items []string, expected string) bool {
	normalizedExpected := normalizeRoutingLabel(expected)
	for _, item := range items {
		if normalizeRoutingLabel(item) == normalizedExpected {
			return true
		}
	}
	return false
}

func normalizeRoutingLabel(raw string) string {
	return strings.TrimSpace(strings.ToLower(raw))
}

func fallbackRoutingLabel(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func multiplyMoney(raw string, qty int) string {
	value, ok := parseMoney(raw)
	if !ok {
		return "TBD"
	}
	return formatMoney(value * float64(qty))
}

func calculateMargin(total, fulfillmentCost, shippingCost, issueCost string) string {
	totalValue, ok := parseMoney(total)
	if !ok {
		return "TBD"
	}
	fulfillmentValue, ok := parseMoney(fulfillmentCost)
	if !ok {
		return "TBD"
	}
	shippingValue, ok := parseMoney(shippingCost)
	if !ok {
		return "TBD"
	}
	issueValue, ok := parseMoney(issueCost)
	if !ok {
		return "TBD"
	}
	return formatMoney(totalValue - fulfillmentValue - shippingValue - issueValue)
}

func normalizeMoney(raw string) (string, error) {
	value, ok := parseMoney(raw)
	if !ok {
		return "", fmt.Errorf("invalid money")
	}
	return formatMoney(value), nil
}

func parseMoney(raw string) (float64, bool) {
	negative := strings.Contains(raw, "-")
	cleaned := strings.Map(func(r rune) rune {
		if (r >= '0' && r <= '9') || r == '.' {
			return r
		}
		return -1
	}, raw)
	if cleaned == "" {
		return 0, false
	}
	var value float64
	if _, err := fmt.Sscanf(cleaned, "%f", &value); err != nil {
		return 0, false
	}
	if negative {
		value = -value
	}
	return value, true
}

func formatMoney(value float64) string {
	return fmt.Sprintf("$%.2f", value)
}
