package routing

import (
	"time"

	"github.com/tuannm99/podzone/pkg/ddd"
)

type DomainEvent = ddd.DomainEvent

type RoutingPartnerSelected struct {
	CandidateID           string
	Partner               string
	EstimatedUnitMargin   string
	EstimatedShippingCost string
	OccurredAt            time.Time
}

var _ DomainEvent = RoutingPartnerSelected{}

func (e RoutingPartnerSelected) EventType() string {
	return "RoutingPartnerSelected"
}

func (e RoutingPartnerSelected) OccurredAtTime() time.Time {
	return e.OccurredAt
}

type RoutingBlocked struct {
	CandidateID string
	ReasonCode  string
	Reason      string
	OccurredAt  time.Time
}

var _ DomainEvent = RoutingBlocked{}

func (e RoutingBlocked) EventType() string {
	return "RoutingBlocked"
}

func (e RoutingBlocked) OccurredAtTime() time.Time {
	return e.OccurredAt
}
