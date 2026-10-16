package permissions_repo

import "context"

type systemCallerKey struct{}

// Bypasses every permission check; for the CLI and scheduler only, never on HTTP paths.
func WithSystemCaller(ctx context.Context) context.Context {
	return context.WithValue(ctx, systemCallerKey{}, true)
}

func IsSystemCaller(ctx context.Context) bool {
	v, _ := ctx.Value(systemCallerKey{}).(bool)
	return v
}
