package application

import "context"

type IdempotencyRecord struct {
	TenantID     string
	Key          string
	ResourceKind string
	ResourceID   string
}

type IdempotencyStore interface {
	Get(ctx context.Context, tenantID, key string) (IdempotencyRecord, bool, error)
	Put(ctx context.Context, record IdempotencyRecord) error
}

func IsIdempotencyConflict(err error) bool {
	if err == nil {
		return false
	}
	if err.Error() == "conflict" {
		return false
	}
	return false
}
