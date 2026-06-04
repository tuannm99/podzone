package routing

import "time"

type DomainEvent interface {
	EventType() string
	OccurredAtTime() time.Time
}

type RoutingPartnerSelected struct {
	CandidateID           string
	Partner               string
	EstimatedUnitMargin   string
	EstimatedShippingCost string
	OccurredAt            time.Time
}

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

func (e RoutingBlocked) EventType() string {
	return "RoutingBlocked"
}

func (e RoutingBlocked) OccurredAtTime() time.Time {
	return e.OccurredAt
}
