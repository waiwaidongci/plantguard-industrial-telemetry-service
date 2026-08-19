package application

import (
	"context"
	"errors"

	shareddomain "github.com/acme/plantguard/internal/shared/domain"
)

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
	return errors.Is(err, shareddomain.ErrConflict)
}
