package settlement

import "time"

type DomainEvent interface {
	EventType() string
	OccurredAtTime() time.Time
}

type SettlementUpdated struct {
	OrderID        string
	Status         string
	RealizedMargin string
	OccurredAt     time.Time
}

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

func (e IssueHandlingUpdated) EventType() string {
	return "IssueHandlingUpdated"
}

func (e IssueHandlingUpdated) OccurredAtTime() time.Time {
	return e.OccurredAt
}
