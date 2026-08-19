package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	shareddomain "github.com/acme/plantguard/internal/shared/domain"
	sharedinfra "github.com/acme/plantguard/internal/shared/infrastructure"
	telemetrydomain "github.com/acme/plantguard/internal/telemetry/domain"
)

func TestStoreBatchHonorsCanceledContext(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := sharedinfra.Migrate(ctx, db, sharedinfra.DatabaseConfig{Driver: "sqlite"}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO tenants(id,name,slug,created_at,updated_at,version) VALUES('tenant-1','Tenant','tenant-1','2024-01-01','2024-01-01',1)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO devices(id,tenant_id,name,protocol,enabled,status,created_at,updated_at,version) VALUES('dev-1','tenant-1','Pump','mqtt',1,'unknown','2024-01-01','2024-01-01',1)`); err != nil {
		t.Fatal(err)
	}

	repo := NewRepository(db)
	canceledCtx, cancel := context.WithCancel(ctx)
	cancel()
	now := time.Unix(1700000000, 0).UTC()
	batch := telemetrydomain.NewBatch("tenant-1", "dev-1", "test", []telemetrydomain.Reading{
		{Metric: "pressure", Value: 1, Timestamp: now},
	}, now)
	t.Run("store batch", func(t *testing.T) {
		if err := repo.StoreBatch(canceledCtx, batch); !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context.Canceled, got %v", err)
		}

		var count int
		if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM telemetry_readings`).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 0 {
			t.Fatalf("canceled batch should not be persisted, found %d readings", count)
		}
	})
	t.Run("list readings", func(t *testing.T) {
		_, _, err := repo.ListReadings(canceledCtx, "tenant-1", "dev-1", shareddomain.PageQuery{Page: 1, PageSize: 20, Filters: map[string]string{}})
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context.Canceled, got %v", err)
		}
	})
}
