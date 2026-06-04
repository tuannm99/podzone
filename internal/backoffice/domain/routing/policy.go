package routing

import (
	"fmt"
	"sort"
	"strings"
	"time"

	catalogentity "github.com/tuannm99/podzone/internal/backoffice/domain/catalog"
)

func NewActivity(
	activityType string,
	actor string,
	message string,
	createdAt time.Time,
	details []RoutedOrderActivityDetail,
) RoutedOrderActivity {
	return RoutedOrderActivity{
		Type:      strings.TrimSpace(activityType),
		Actor:     strings.TrimSpace(actor),
		Message:   strings.TrimSpace(message),
		Details:   details,
		CreatedAt: createdAt,
	}
}

func ActivityDetails(pairs ...string) []RoutedOrderActivityDetail {
	details := make([]RoutedOrderActivityDetail, 0, len(pairs)/2)
	for i := 0; i+1 < len(pairs); i += 2 {
		key := strings.TrimSpace(pairs[i])
		value := strings.TrimSpace(pairs[i+1])
		if key == "" || value == "" {
			continue
		}
		details = append(details, RoutedOrderActivityDetail{Key: key, Value: value})
	}
	return details
}

func RoutingTimelineEntry(status string, partner string) string {
	switch status {
	case RoutedOrderStatusInProduction:
		return fmt.Sprintf("Sent to %s for POD production", partner)
	case RoutedOrderStatusShipped:
		return fmt.Sprintf("Marked as shipped from %s", partner)
	default:
		return fmt.Sprintf("Queued for %s", partner)
	}
}

func BuildRoutingRecommendation(
	candidate *catalogentity.ProductSetupCandidate,
	partners []PartnerRoutingProfile,
	productType string,
	shipRegion string,
	preferredPartner string,
) *RoutedOrderRecommendation {
	options := make([]RoutingPartnerOption, 0, len(partners))

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
		options = append(options, RoutingPartnerOption{
			Partner:                  partner,
			Eligible:                 eligible,
			Reason:                   describeRoutingOption(reason, estimatedUnitMargin),
			EstimatedFulfillmentCost: estimatedFulfillmentCost,
			EstimatedShippingCost:    estimatedShippingCost,
			EstimatedUnitMargin:      estimatedUnitMargin,
		})
	}

	sort.SliceStable(options, func(i, j int) bool {
		left := options[i]
		right := options[j]
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

	decision, err := NewRoutingDecision(
		candidate.ID,
		candidate.Title,
		candidate.Partner,
		productType,
		shipRegion,
		options,
	)
	if err != nil {
		return &RoutedOrderRecommendation{
			CandidateID:       strings.TrimSpace(candidate.ID),
			ProductTitle:      strings.TrimSpace(candidate.Title),
			CandidatePartner:  strings.TrimSpace(candidate.Partner),
			ProductType:       NormalizeRoutingLabel(productType),
			ShipRegion:        NormalizeRoutingLabel(shipRegion),
			BlockedReasonCode: "invalid_candidate",
			BlockedReason:     err.Error(),
			Summary:           err.Error(),
			Options:           options,
		}
	}
	now := time.Now().UTC()
	preferredPartner = strings.TrimSpace(preferredPartner)
	if preferredPartner != "" {
		if decision.SelectPreferred(preferredPartner, now) {
			return recommendationFromDecision(decision)
		}
	}

	if decision.SelectBestAvailable(now) {
		return recommendationFromDecision(decision)
	}

	if hasEligibleOptions(options) {
		decision.Block(
			"negative_margin",
			"all eligible partners have negative expected margin",
			fmt.Sprintf(
				"No auto-route partner selected for %s in %s because all eligible options have negative expected margin.",
				fallbackRoutingLabel(productType, "the requested product"),
				fallbackRoutingLabel(shipRegion, "the requested region"),
			),
			now,
		)
		return recommendationFromDecision(decision)
	}

	reasonCode, reason := deriveRoutingBlockReason(options)
	decision.Block(
		reasonCode,
		reason,
		fmt.Sprintf(
			"No active partner matches %s in %s.",
			fallbackRoutingLabel(productType, "the requested product"),
			fallbackRoutingLabel(shipRegion, "the requested region"),
		),
		now,
	)
	return recommendationFromDecision(decision)
}

func FindSelectedRoutingOption(
	recommendation *RoutedOrderRecommendation,
) *RoutingPartnerOption {
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

func NormalizeRoutingLabel(raw string) string {
	return strings.TrimSpace(strings.ToLower(raw))
}

func evaluatePartnerEligibility(
	partner PartnerRoutingProfile,
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

func routingOptionRank(option RoutingPartnerOption) int {
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

func hasEligibleOptions(options []RoutingPartnerOption) bool {
	for _, option := range options {
		if option.Eligible {
			return true
		}
	}
	return false
}

func summarizeMarginAwareSelection(candidatePartner string, option RoutingPartnerOption) string {
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

func describeRoutingOption(baseReason string, estimatedUnitMargin string) string {
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

func deriveRoutingBlockReason(options []RoutingPartnerOption) (string, string) {
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

func estimateFulfillmentCost(candidateBaseCost string, partnerBaseCost string) string {
	if strings.TrimSpace(partnerBaseCost) != "" {
		return partnerBaseCost
	}
	if strings.TrimSpace(candidateBaseCost) != "" {
		return candidateBaseCost
	}
	return "TBD"
}

func estimateShippingCost(rules []PartnerShippingCostRule, shipRegion string) string {
	normalizedRegion := NormalizeRoutingLabel(shipRegion)
	for _, rule := range rules {
		if NormalizeRoutingLabel(rule.Region) == normalizedRegion && strings.TrimSpace(rule.Cost) != "" {
			return rule.Cost
		}
	}
	return "$0.00"
}

func estimateUnitMargin(
	retailPrice string,
	candidateBaseCost string,
	partner PartnerRoutingProfile,
	shipRegion string,
) string {
	fulfillmentCost := estimateFulfillmentCost(candidateBaseCost, partner.BaseFulfillmentCost)
	shippingCost := estimateShippingCost(partner.ShippingCostRules, shipRegion)
	return CalculateMargin(retailPrice, fulfillmentCost, shippingCost, "$0.00")
}

func containsNormalized(items []string, expected string) bool {
	normalizedExpected := NormalizeRoutingLabel(expected)
	for _, item := range items {
		if NormalizeRoutingLabel(item) == normalizedExpected {
			return true
		}
	}
	return false
}

func fallbackRoutingLabel(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
