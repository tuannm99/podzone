package routing

import (
	"fmt"
	"strings"
	"time"

	"github.com/tuannm99/podzone/pkg/ddd"
)

type RoutingDecisionSnapshot struct {
	CandidateID       string
	ProductTitle      string
	CandidatePartner  string
	ProductType       string
	ShipRegion        string
	SelectedPartner   string
	BlockedReasonCode string
	BlockedReason     string
	Summary           string
	Options           []RoutingPartnerOption
}

type RoutingDecision struct {
	aggregate         ddd.AggregateBase
	candidateID       string
	productTitle      string
	candidatePartner  string
	productType       string
	shipRegion        string
	selectedPartner   string
	blockedReasonCode string
	blockedReason     string
	summary           string
	options           []RoutingPartnerOption
}

var _ ddd.AggregateRoot = (*RoutingDecision)(nil)

func NewRoutingDecision(
	candidateID string,
	productTitle string,
	candidatePartner string,
	productType string,
	shipRegion string,
	options []RoutingPartnerOption,
) (*RoutingDecision, error) {
	candidateID = strings.TrimSpace(candidateID)
	if candidateID == "" {
		return nil, ErrRoutingCandidateRequired
	}
	aggregate, err := newAggregate(candidateID)
	if err != nil {
		return nil, err
	}
	return &RoutingDecision{
		aggregate:        aggregate,
		candidateID:      candidateID,
		productTitle:     strings.TrimSpace(productTitle),
		candidatePartner: strings.TrimSpace(candidatePartner),
		productType:      NormalizeRoutingLabel(productType),
		shipRegion:       NormalizeRoutingLabel(shipRegion),
		options:          append([]RoutingPartnerOption(nil), options...),
	}, nil
}

func (d *RoutingDecision) SelectPreferred(preferredPartner string, now time.Time) bool {
	preferredPartner = strings.TrimSpace(preferredPartner)
	if preferredPartner == "" {
		return false
	}
	for _, option := range d.options {
		if !option.Eligible {
			continue
		}
		if strings.EqualFold(option.Partner.Name, preferredPartner) ||
			strings.EqualFold(option.Partner.Code, preferredPartner) {
			d.selectPartner(option, preferredPartnerSummary(option, d.productType, d.shipRegion), now)
			return true
		}
	}
	d.summary = fmt.Sprintf(
		"Preferred partner %s is not eligible for %s in %s.",
		preferredPartner,
		fallbackRoutingLabel(d.productType, "any product"),
		fallbackRoutingLabel(d.shipRegion, "any region"),
	)
	return false
}

func (d *RoutingDecision) SelectBestAvailable(now time.Time) bool {
	for _, option := range d.options {
		if option.Eligible && routingOptionRank(option) <= routingOptionRankUnknownMargin {
			d.selectPartner(option, summarizeMarginAwareSelection(d.candidatePartner, option), now)
			return true
		}
	}
	return false
}

func (d *RoutingDecision) Block(reasonCode string, reason string, summary string, now time.Time) {
	d.selectedPartner = ""
	d.blockedReasonCode = strings.TrimSpace(reasonCode)
	d.blockedReason = strings.TrimSpace(reason)
	d.summary = strings.TrimSpace(summary)
	d.record(RoutingBlocked{
		CandidateID: d.candidateID,
		ReasonCode:  d.blockedReasonCode,
		Reason:      d.blockedReason,
		OccurredAt:  now.UTC(),
	})
}

func (d *RoutingDecision) Snapshot() RoutingDecisionSnapshot {
	return RoutingDecisionSnapshot{
		CandidateID:       d.candidateID,
		ProductTitle:      d.productTitle,
		CandidatePartner:  d.candidatePartner,
		ProductType:       d.productType,
		ShipRegion:        d.shipRegion,
		SelectedPartner:   d.selectedPartner,
		BlockedReasonCode: d.blockedReasonCode,
		BlockedReason:     d.blockedReason,
		Summary:           d.summary,
		Options:           append([]RoutingPartnerOption(nil), d.options...),
	}
}

func (d *RoutingDecision) AggregateID() ddd.ID {
	if d == nil {
		return ""
	}
	return d.aggregate.AggregateID()
}

func (d *RoutingDecision) AggregateVersion() ddd.Version {
	if d == nil {
		return 0
	}
	return d.aggregate.AggregateVersion()
}

func (d *RoutingDecision) PullEvents() []DomainEvent {
	if d == nil {
		return nil
	}
	return d.aggregate.PullEvents()
}

func (d *RoutingDecision) selectPartner(option RoutingPartnerOption, summary string, now time.Time) {
	d.selectedPartner = option.Partner.Name
	d.blockedReasonCode = ""
	d.blockedReason = ""
	d.summary = summary
	d.record(RoutingPartnerSelected{
		CandidateID:           d.candidateID,
		Partner:               d.selectedPartner,
		EstimatedUnitMargin:   option.EstimatedUnitMargin,
		EstimatedShippingCost: option.EstimatedShippingCost,
		OccurredAt:            now.UTC(),
	})
}

func (d *RoutingDecision) record(event DomainEvent) {
	d.aggregate.RecordEvent(event)
}

func newAggregate(rawID string) (ddd.AggregateBase, error) {
	id, err := ddd.ParseID(rawID)
	if err != nil {
		return ddd.AggregateBase{}, err
	}
	return ddd.NewAggregateBase(id, 0)
}

func preferredPartnerSummary(option RoutingPartnerOption, productType string, shipRegion string) string {
	return fmt.Sprintf(
		"Preferred partner %s is eligible for %s in %s with expected unit margin %s.",
		option.Partner.Name,
		fallbackRoutingLabel(productType, "any product"),
		fallbackRoutingLabel(shipRegion, "any region"),
		option.EstimatedUnitMargin,
	)
}

func recommendationFromDecision(decision *RoutingDecision) *RoutedOrderRecommendation {
	snapshot := decision.Snapshot()
	return &RoutedOrderRecommendation{
		CandidateID:       snapshot.CandidateID,
		ProductTitle:      snapshot.ProductTitle,
		CandidatePartner:  snapshot.CandidatePartner,
		ProductType:       snapshot.ProductType,
		ShipRegion:        snapshot.ShipRegion,
		SelectedPartner:   snapshot.SelectedPartner,
		BlockedReasonCode: snapshot.BlockedReasonCode,
		BlockedReason:     snapshot.BlockedReason,
		Summary:           snapshot.Summary,
		Options:           snapshot.Options,
	}
}
