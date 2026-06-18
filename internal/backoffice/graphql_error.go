package backoffice

import (
	"context"
	"errors"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

const (
	graphQLErrorCodeUnauthenticated = "UNAUTHENTICATED"
	graphQLErrorCodeForbidden       = "FORBIDDEN"
	graphQLErrorCodeNotFound        = "NOT_FOUND"
	graphQLErrorCodeFailed          = "FAILED_PRECONDITION"
	graphQLErrorCodeInternal        = "INTERNAL"
)

func graphQLErrorResponse(ctx context.Context, err error) *graphql.Response {
	return &graphql.Response{Errors: gqlerror.List{graphQLError(ctx, err)}}
}

func graphQLError(ctx context.Context, err error) *gqlerror.Error {
	if err == nil {
		err = errors.New("request failed")
	}

	var gqlErr *gqlerror.Error
	if ctx == nil {
		gqlErr = &gqlerror.Error{Message: err.Error()}
	} else {
		gqlErr = graphql.DefaultErrorPresenter(ctx, err)
	}
	if gqlErr.Extensions == nil {
		gqlErr.Extensions = map[string]any{}
	}

	var permissionErr *PermissionDeniedError
	var permissionMappingErr *PermissionMappingError
	switch {
	case errors.As(err, &permissionErr):
		gqlErr.Message = permissionErr.Error()
		gqlErr.Extensions["code"] = graphQLErrorCodeForbidden
		gqlErr.Extensions["permission"] = permissionErr.Permission
		gqlErr.Extensions["resource"] = permissionErr.Resource
	case errors.As(err, &permissionMappingErr):
		gqlErr.Message = permissionMappingErr.Error()
		gqlErr.Extensions["code"] = graphQLErrorCodeInternal
		gqlErr.Extensions["field"] = permissionMappingErr.Object + "." + permissionMappingErr.Field
	case isUnauthenticatedError(err):
		gqlErr.Extensions["code"] = graphQLErrorCodeUnauthenticated
	case isNotFoundError(err):
		gqlErr.Extensions["code"] = graphQLErrorCodeNotFound
	case isFailedPreconditionError(err):
		gqlErr.Extensions["code"] = graphQLErrorCodeFailed
	default:
		gqlErr.Extensions["code"] = graphQLErrorCodeInternal
	}
	return gqlErr
}

func isUnauthenticatedError(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "authorization") ||
		strings.Contains(msg, "invalid authorization") ||
		strings.Contains(msg, "session")
}

func isNotFoundError(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "not found")
}

func isFailedPreconditionError(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "inactive") ||
		strings.Contains(msg, "mismatch") ||
		strings.Contains(msg, "bootstrap failed")
}
