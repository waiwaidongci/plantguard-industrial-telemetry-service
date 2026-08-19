package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	sharedapplication "github.com/acme/plantguard/internal/shared/application"
	shareddomain "github.com/acme/plantguard/internal/shared/domain"
)

func TestIdempotencyPutPreservesConflict(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := Migrate(ctx, db, DatabaseConfig{Driver: "sqlite"}); err != nil {
		t.Fatal(err)
	}
	store := NewIdempotencyStore(db)
	record := sharedapplication.IdempotencyRecord{
		TenantID:     "tenant-1",
		Key:          "request-1",
		ResourceKind: "device",
		ResourceID:   "device-1",
	}
	if err := store.Put(ctx, record); err != nil {
		t.Fatal(err)
	}
	record.ResourceID = "device-2"
	err = store.Put(ctx, record)
	if !errors.Is(err, shareddomain.ErrConflict) {
		t.Fatalf("expected conflict error to be preserved, got %T: %v", err, err)
	}
}
