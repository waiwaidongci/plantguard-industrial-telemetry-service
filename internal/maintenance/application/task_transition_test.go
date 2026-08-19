package application

import (
	"context"
	"errors"
	"testing"
	"time"

	maintenancedomain "github.com/acme/plantguard/internal/maintenance/domain"
	shareddomain "github.com/acme/plantguard/internal/shared/domain"
)

type retryingTaskRepo struct {
	updated maintenancedomain.Task
}

func (r *retryingTaskRepo) Create(context.Context, maintenancedomain.Task) error { return nil }
func (r *retryingTaskRepo) GetByID(context.Context, string, string) (maintenancedomain.Task, error) {
	return maintenancedomain.Task{
		ID:        "task-1",
		TenantID:  "tenant-1",
		Status:    "retrying",
		CreatedAt: time.Unix(1700000000, 0).UTC(),
		UpdatedAt: time.Unix(1700000000, 0).UTC(),
		Version:   1,
	}, nil
}
func (r *retryingTaskRepo) Update(_ context.Context, task maintenancedomain.Task) error {
	r.updated = task
	return nil
}
func (retryingTaskRepo) List(context.Context, string, shareddomain.PageQuery) ([]maintenancedomain.Task, int64, error) {
	return nil, 0, nil
}
func (retryingTaskRepo) ListOpenByDevice(context.Context, string, string) ([]maintenancedomain.Task, error) {
	return nil, nil
}

type transitionClock struct{}

func (transitionClock) Now(context.Context) time.Time { return time.Unix(1700000001, 0).UTC() }

func TestRetryingTaskTransitions(t *testing.T) {
	repo := &retryingTaskRepo{}
	svc := NewService(nil, repo, nil, transitionClock{})
	t.Run("complete retrying task", func(t *testing.T) {
		task, err := svc.TransitionTask(context.Background(), "tenant-1", "task-1", TransitionTaskInput{Status: "completed", Notes: "done"})
		if err != nil {
			t.Fatalf("expected retrying task to complete, got %v", err)
		}
		if task.Status != "completed" {
			t.Fatalf("expected status completed, got %q", task.Status)
		}
		if task.CompletedAt.IsZero() {
			t.Fatal("expected CompletedAt to be set when task completes")
		}
		if repo.updated.Version != 2 {
			t.Fatalf("expected task version to increment to 2, got %d", repo.updated.Version)
		}
		if repo.updated.Notes != "done" {
			t.Fatalf("expected notes to be persisted, got %q", repo.updated.Notes)
		}
		if repo.updated.UpdatedAt.IsZero() {
			t.Fatal("expected UpdatedAt to be set")
		}
	})
	t.Run("version mismatch", func(t *testing.T) {
		_, err := svc.TransitionTask(context.Background(), "tenant-1", "task-1", TransitionTaskInput{Status: "completed", Version: 99})
		if !errors.Is(err, shareddomain.ErrPrecondition) {
			t.Fatalf("expected precondition error, got %T: %v", err, err)
		}
	})
}
