package catalog

import "time"

type DomainEvent interface {
	EventType() string
	OccurredAtTime() time.Time
}

type ProductSetupDraftCreated struct {
	DraftID    string
	StoreID    string
	Name       string
	OccurredAt time.Time
}

func (e ProductSetupDraftCreated) EventType() string {
	return "ProductSetupDraftCreated"
}

func (e ProductSetupDraftCreated) OccurredAtTime() time.Time {
	return e.OccurredAt
}

type ProductSetupCandidatePromoted struct {
	CandidateID string
	DraftID     string
	StoreID     string
	OccurredAt  time.Time
}

func (e ProductSetupCandidatePromoted) EventType() string {
	return "ProductSetupCandidatePromoted"
}

func (e ProductSetupCandidatePromoted) OccurredAtTime() time.Time {
	return e.OccurredAt
}
