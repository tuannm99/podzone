package outputport

import "context"

type AccessAuthorizer interface {
	AuthorizeInfrastructureRead(ctx context.Context, requestedBy string) error
	AuthorizeInfrastructureManage(ctx context.Context, requestedBy string) error
}
