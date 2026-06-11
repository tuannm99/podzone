package catalog

import (
	"time"

	"github.com/tuannm99/podzone/pkg/ddd"
)

type DomainEvent = ddd.DomainEvent

type ProductSetupDraftCreated struct {
	DraftID    string
	StoreID    string
	Name       string
	OccurredAt time.Time
}

var _ DomainEvent = ProductSetupDraftCreated{}

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

var _ DomainEvent = ProductSetupCandidatePromoted{}

func (e ProductSetupCandidatePromoted) EventType() string {
	return "ProductSetupCandidatePromoted"
}

func (e ProductSetupCandidatePromoted) OccurredAtTime() time.Time {
	return e.OccurredAt
}
