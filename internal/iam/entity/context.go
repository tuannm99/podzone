package entity

import "context"

type (
	sessionPolicyContextKey struct{}
	assumedRoleContextKey   struct{}
	sessionTagsContextKey   struct{}
)

func WithSessionPolicyStatements(ctx context.Context, statements []PolicyStatement) context.Context {
	if len(statements) == 0 {
		return ctx
	}
	cloned := make([]PolicyStatement, len(statements))
	copy(cloned, statements)
	return context.WithValue(ctx, sessionPolicyContextKey{}, cloned)
}

func GetSessionPolicyStatements(ctx context.Context) []PolicyStatement {
	statements, ok := ctx.Value(sessionPolicyContextKey{}).([]PolicyStatement)
	if !ok || len(statements) == 0 {
		return nil
	}
	cloned := make([]PolicyStatement, len(statements))
	copy(cloned, statements)
	return cloned
}

func WithAssumedRole(ctx context.Context, assumedRole AssumedRole) context.Context {
	return context.WithValue(ctx, assumedRoleContextKey{}, assumedRole)
}

func GetAssumedRole(ctx context.Context) (*AssumedRole, bool) {
	assumedRole, ok := ctx.Value(assumedRoleContextKey{}).(AssumedRole)
	if !ok {
		return nil, false
	}
	return &assumedRole, true
}

func WithSessionTags(ctx context.Context, tags map[string]string) context.Context {
	if len(tags) == 0 {
		return ctx
	}
	cloned := make(map[string]string, len(tags))
	for k, v := range tags {
		cloned[k] = v
	}
	return context.WithValue(ctx, sessionTagsContextKey{}, cloned)
}

func GetSessionTags(ctx context.Context) map[string]string {
	tags, ok := ctx.Value(sessionTagsContextKey{}).(map[string]string)
	if !ok || len(tags) == 0 {
		return nil
	}
	cloned := make(map[string]string, len(tags))
	for k, v := range tags {
		cloned[k] = v
	}
	return cloned
}
