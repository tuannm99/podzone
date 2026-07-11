package iamprojection

import (
	"context"
	"errors"

	"github.com/tuannm99/podzone/internal/auth/domain/outputport"
	"github.com/tuannm99/podzone/pkg/messaging"
)

func NewHandler(repo outputport.IAMProjectionRepository) (messaging.Handler, error) {
	registry, err := messaging.NewRegistry(
		NewTenantCreatedHandler(repo),
		NewTenantMemberAddedHandler(repo),
	)
	if err != nil {
		return nil, err
	}

	// This projection only cares about a subset of iam.* events; other event
	// types on the same topic are expected and must not error/retry/dead-letter.
	return messaging.HandlerFunc(func(ctx context.Context, msg messaging.Envelope) error {
		err := registry.Handle(ctx, msg)
		if errors.Is(err, messaging.ErrHandlerNotFound) {
			return nil
		}
		return err
	}), nil
}
