package kvstorebridge

import (
	"github.com/tuannm99/podzone/internal/onboarding/infrastructure/messaging/publisher"
	"github.com/tuannm99/podzone/pkg/messaging"
)

func NewRegistry(pub *publisher.KVStorePublisher) (*messaging.Registry, error) {
	return messaging.NewRegistry(
		NewPublishHandler(pub),
		NewDeleteHandler(pub),
	)
}
