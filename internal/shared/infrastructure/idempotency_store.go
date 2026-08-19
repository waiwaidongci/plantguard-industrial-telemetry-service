package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	sharedapplication "github.com/acme/plantguard/internal/shared/application"
)

type IdempotencyStore struct {
	db *sql.DB
}

func NewIdempotencyStore(db *sql.DB) *IdempotencyStore {
	return &IdempotencyStore{db: db}
}

func (s *IdempotencyStore) Get(ctx context.Context, tenantID, key string) (sharedapplication.IdempotencyRecord, bool, error) {
	var record sharedapplication.IdempotencyRecord
	err := s.db.QueryRowContext(ctx, `SELECT tenant_id, key, resource_kind, resource_id FROM idempotency_keys WHERE tenant_id=? AND key=?`, tenantID, key).
		Scan(&record.TenantID, &record.Key, &record.ResourceKind, &record.ResourceID)
	if err == sql.ErrNoRows {
		return record, false, nil
	}
	if err != nil {
		return record, false, err
	}
	return record, true, nil
}

func (s *IdempotencyStore) Put(ctx context.Context, record sharedapplication.IdempotencyRecord) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO idempotency_keys(tenant_id, key, resource_kind, resource_id, created_at) VALUES(?, ?, ?, ?, ?)`,
		record.TenantID, record.Key, record.ResourceKind, record.ResourceID, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("insert idempotency key: %v", err)
	}
	return nil
}
