package operations

import (
	"context"

	"github.com/tuannm99/podzone/pkg/ddd"
)

func collectDomainEvents[T ddd.DomainEvent](domainEvents []T) []ddd.DomainEvent {
	out := make([]ddd.DomainEvent, 0, len(domainEvents))
	for _, event := range domainEvents {
		out = append(out, event)
	}
	return out
}

func dispatchDomainEvents(
	ctx context.Context,
	dispatcher ddd.EventDispatcher,
	domainEvents []ddd.DomainEvent,
) error {
	if dispatcher == nil || len(domainEvents) == 0 {
		return nil
	}
	return dispatcher.Dispatch(ctx, domainEvents)
}
