package settlement

import (
	"time"

	"github.com/tuannm99/podzone/pkg/ddd"
)

type DomainEvent = ddd.DomainEvent

type SettlementUpdated struct {
	OrderID        string
	Status         string
	RealizedMargin string
	OccurredAt     time.Time
}

var _ DomainEvent = SettlementUpdated{}

func (e SettlementUpdated) EventType() string {
	return "SettlementUpdated"
}

func (e SettlementUpdated) OccurredAtTime() time.Time {
	return e.OccurredAt
}

type IssueHandlingUpdated struct {
	OrderID         string
	IssueResolution string
	IssueCost       string
	RealizedMargin  string
	OccurredAt      time.Time
}

var _ DomainEvent = IssueHandlingUpdated{}

func (e IssueHandlingUpdated) EventType() string {
	return "IssueHandlingUpdated"
}

func (e IssueHandlingUpdated) OccurredAtTime() time.Time {
	return e.OccurredAt
}
