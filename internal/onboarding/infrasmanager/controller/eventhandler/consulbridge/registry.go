package consulbridge

import (
	"github.com/tuannm99/podzone/internal/onboarding/infrasmanager/core/publisher"
	"github.com/tuannm99/podzone/pkg/messaging"
)

func NewRegistry(pub *publisher.ConsulPublisher) (*messaging.Registry, error) {
	return messaging.NewRegistry(
		NewConsulPublishHandler(pub),
		NewConsulDeleteHandler(pub),
	)
}
