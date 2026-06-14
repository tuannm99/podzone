package identity

import "context"

type contextKey string

const trustedServiceKey contextKey = "trusted_service"

func WithTrustedService(ctx context.Context) context.Context {
	return context.WithValue(ctx, trustedServiceKey, true)
}

func IsTrustedService(ctx context.Context) bool {
	trusted, _ := ctx.Value(trustedServiceKey).(bool)
	return trusted
}
