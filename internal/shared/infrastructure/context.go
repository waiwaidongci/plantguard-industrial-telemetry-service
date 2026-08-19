package infrastructure

import "context"

type tenantKey struct{}

func WithTenantID(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, tenantKey{}, tenantID)
}

func TenantIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(tenantKey{}).(string)
	return id
}
