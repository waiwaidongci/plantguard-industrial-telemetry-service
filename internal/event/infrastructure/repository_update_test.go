package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	eventdomain "github.com/acme/plantguard/internal/event/domain"
	shareddomain "github.com/acme/plantguard/internal/shared/domain"
	sharedinfra "github.com/acme/plantguard/internal/shared/infrastructure"
)

func TestEventRepositoryUpdatePreservesPrecondition(t *testing.T) {
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
	now := time.Unix(1700000000, 0).UTC()
	event := eventdomain.NewEvent("tenant-1", "dev-1", "", "alert", "warning", "hello", nil, now, now)
	repo := NewRepository(db)
	if err := repo.Create(ctx, event); err != nil {
		t.Fatal(err)
	}
	event.Version = 2
	if err := repo.Update(ctx, event); err != nil {
		t.Fatal(err)
	}
	err = repo.Update(ctx, event)
	if !errors.Is(err, shareddomain.ErrPrecondition) {
		t.Fatalf("expected precondition error, got %T: %v", err, err)
	}
}
