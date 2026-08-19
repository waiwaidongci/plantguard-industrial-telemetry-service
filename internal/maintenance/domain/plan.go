package domain

import (
	"context"
	"time"

	shareddomain "github.com/acme/plantguard/internal/shared/domain"
)

type Plan struct {
	ID            string    `json:"id"`
	TenantID      string    `json:"tenant_id"`
	DeviceID      string    `json:"device_id"`
	Name          string    `json:"name"`
	TriggerType   string    `json:"trigger_type"`
	IntervalHours int       `json:"interval_hours"`
	CalendarSpec  string    `json:"calendar_spec"`
	EventType     string    `json:"event_type"`
	Status        string    `json:"status"`
	NextDueAt     time.Time `json:"next_due_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Version       int64     `json:"version"`
}

func NewPlan(tenantID, deviceID, name, triggerType string, intervalHours int, calendarSpec, eventType string, now time.Time) Plan {
	var nextDueAt time.Time
	if triggerType == "event" {
		nextDueAt = time.Time{}
	} else if intervalHours > 0 {
		nextDueAt = now.Add(time.Duration(intervalHours) * time.Hour)
	} else {
		nextDueAt = now.Add(24 * time.Hour)
	}
	return Plan{
		ID:            shareddomain.NewID("pln"),
		TenantID:      tenantID,
		DeviceID:      deviceID,
		Name:          name,
		TriggerType:   triggerType,
		IntervalHours: intervalHours,
		CalendarSpec:  calendarSpec,
		EventType:     eventType,
		Status:        "active",
		NextDueAt:     nextDueAt,
		CreatedAt:     now,
		UpdatedAt:     now,
		Version:       1,
	}
}

type PlanRepository interface {
	Create(context.Context, Plan) error
	GetByID(context.Context, string, string) (Plan, error)
	Update(context.Context, Plan) error
	List(context.Context, string, shareddomain.PageQuery) ([]Plan, int64, error)
	ListActiveDue(context.Context, time.Time) ([]Plan, error)
}
