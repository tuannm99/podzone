package interactor

import (
	"context"
	"fmt"

	"github.com/tuannm99/podzone/internal/iam/domain/entity"
	"github.com/tuannm99/podzone/pkg/collection"
)

func (s *interactor) ListDirectoryUsers(
	ctx context.Context,
	query collection.Query,
) (collection.Page[entity.User], error) {
	if s.userDirectory == nil {
		return collection.Page[entity.User]{}, fmt.Errorf("iam: user directory is unavailable")
	}
	return s.userDirectory.List(ctx, query.Normalize())
}

func (s *interactor) ListPermissions(
	ctx context.Context,
	query collection.Query,
) (collection.Page[entity.Permission], error) {
	return s.roleQueries.ListPermissions(ctx, query.Normalize())
}
