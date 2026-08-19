package domain

import (
	"context"
	"time"

	shareddomain "github.com/acme/plantguard/internal/shared/domain"
)

type Task struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	PlanID      string    `json:"plan_id"`
	DeviceID    string    `json:"device_id"`
	Status      string    `json:"status"`
	DueAt       time.Time `json:"due_at"`
	CompletedAt time.Time `json:"completed_at"`
	Notes       string    `json:"notes"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Version     int64     `json:"version"`
}

func NewTask(tenantID, planID, deviceID string, dueAt, now time.Time) Task {
	return Task{
		ID:        shareddomain.NewID("tsk"),
		TenantID:  tenantID,
		PlanID:    planID,
		DeviceID:  deviceID,
		Status:    "open",
		DueAt:     dueAt,
		CreatedAt: now,
		UpdatedAt: now,
		Version:   1,
	}
}

func (t *Task) CanTransitionTo(status string) bool {
	switch status {
	case "in_progress":
		return t.Status == "open"
	case "completed":
		return t.Status == "open" || t.Status == "in_progress"
	case "cancelled":
		return t.Status == "open" || t.Status == "in_progress"
	default:
		return false
	}
}

type TaskRepository interface {
	Create(context.Context, Task) error
	GetByID(context.Context, string, string) (Task, error)
	Update(context.Context, Task) error
	List(context.Context, string, shareddomain.PageQuery) ([]Task, int64, error)
	ListOpenByDevice(context.Context, string, string) ([]Task, error)
}
