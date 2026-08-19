package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	devicedomain "github.com/acme/plantguard/internal/device/domain"
	shareddomain "github.com/acme/plantguard/internal/shared/domain"
	sharedinfra "github.com/acme/plantguard/internal/shared/infrastructure"
)

func TestModelRepositoryPreservesNotFound(t *testing.T) {
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
	repo := NewModelRepository(db)
	t.Run("get missing model", func(t *testing.T) {
		_, err := repo.GetByID(ctx, "tenant-1", "missing")
		if !errors.Is(err, shareddomain.ErrNotFound) {
			t.Fatalf("expected not-found error to be preserved, got %T: %v", err, err)
		}
	})
	t.Run("update missing model", func(t *testing.T) {
		model := devicedomain.NewDeviceModel("tenant-1", "missing", nil, time.Unix(1700000000, 0).UTC())
		model.ID = "missing"
		err := repo.Update(ctx, model)
		if !errors.Is(err, shareddomain.ErrPrecondition) {
			t.Fatalf("expected precondition error to be preserved, got %T: %v", err, err)
		}
	})
}
