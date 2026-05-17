package domain

import "context"

type sessionPolicyContextKey struct{}
type assumedRoleContextKey struct{}

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
