package outputport

import "context"

type AccountBootstrapper interface {
	EnsureRootOrganization(
		ctx context.Context,
		accessToken string,
		userID uint,
		username string,
	) error
}
