package infrastructure

import (
	"context"
	"database/sql"
	"testing"

	sharedinfra "github.com/acme/plantguard/internal/shared/infrastructure"
)

func TestListOpenByDeviceIncludesRetrying(t *testing.T) {
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
	if _, err := db.ExecContext(ctx, `INSERT INTO maintenance_plans(id,tenant_id,device_id,name,trigger_type,status,created_at,updated_at,version) VALUES('plan-1','tenant-1','dev-1','Plan','calendar','active','2024-01-01','2024-01-01',1)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO maintenance_tasks(id,tenant_id,plan_id,device_id,status,created_at,updated_at,version) VALUES('task-1','tenant-1','plan-1','dev-1','retrying','2024-01-01','2024-01-01',1)`); err != nil {
		t.Fatal(err)
	}

	repo := NewTaskRepository(db)
	tasks, err := repo.ListOpenByDevice(ctx, "tenant-1", "dev-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 1 || tasks[0].Status != "retrying" {
		t.Fatalf("expected retrying task to be visible, got %#v", tasks)
	}
}
