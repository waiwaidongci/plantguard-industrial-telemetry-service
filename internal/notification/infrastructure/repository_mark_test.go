package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	sharedinfra "github.com/acme/plantguard/internal/shared/infrastructure"
)

func TestNotificationRepositoryMarkSentPropagatesError(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := sharedinfra.Migrate(ctx, db, sharedinfra.DatabaseConfig{Driver: "sqlite"}); err != nil {
		t.Fatal(err)
	}
	db.Close()

	repo := NewRepository(db)
	err = repo.MarkSent(ctx, "tenant-1", "notification-1", time.Unix(1700000000, 0).UTC())
	if err == nil {
		t.Fatal("expected MarkSent error to be propagated")
	}
	if !errors.Is(err, sql.ErrTxDone) && !errors.Is(err, context.Canceled) && !errors.Is(err, sql.ErrConnDone) {
		t.Logf("got error: %v", err)
	}
}

func TestNotificationRepositoryMarkFailedPropagatesError(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := sharedinfra.Migrate(ctx, db, sharedinfra.DatabaseConfig{Driver: "sqlite"}); err != nil {
		t.Fatal(err)
	}
	db.Close()

	repo := NewRepository(db)
	err = repo.MarkFailed(ctx, "tenant-1", "notification-1", "boom")
	if err == nil {
		t.Fatal("expected MarkFailed error to be propagated")
	}
}
