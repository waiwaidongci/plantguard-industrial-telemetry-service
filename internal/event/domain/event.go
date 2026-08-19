package domain

import (
	"context"
	"encoding/json"
	"time"

	shareddomain "github.com/acme/plantguard/internal/shared/domain"
)

type Event struct {
	ID             string         `json:"id"`
	TenantID       string         `json:"tenant_id"`
	DeviceID       string         `json:"device_id"`
	RuleID         string         `json:"rule_id"`
	Type           string         `json:"type"`
	Severity       string         `json:"severity"`
	Message        string         `json:"message"`
	Data           map[string]any `json:"data"`
	Status         string         `json:"status"`
	OccurredAt     time.Time      `json:"occurred_at"`
	AcknowledgedBy string         `json:"acknowledged_by"`
	AcknowledgedAt time.Time      `json:"acknowledged_at"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	Version        int64          `json:"version"`
}

func NewEvent(tenantID, deviceID, ruleID, eventType, severity, message string, data map[string]any, occurredAt, now time.Time) Event {
	if data == nil {
		data = map[string]any{}
	}
	return Event{
		ID:         shareddomain.NewID("evt"),
		TenantID:   tenantID,
		DeviceID:   deviceID,
		RuleID:     ruleID,
		Type:       eventType,
		Severity:   severity,
		Message:    message,
		Data:       data,
		Status:     "open",
		OccurredAt: occurredAt,
		CreatedAt:  now,
		UpdatedAt:  now,
		Version:    1,
	}
}

func (e Event) MarshalData() (string, error) {
	raw, err := json.Marshal(e.Data)
	return string(raw), err
}

func UnmarshalData(raw string) map[string]any {
	data := map[string]any{}
	_ = json.Unmarshal([]byte(raw), &data)
	return data
}

type EventRepository interface {
	Create(context.Context, Event) error
	GetByID(context.Context, string, string) (Event, error)
	Update(context.Context, Event) error
	List(context.Context, string, shareddomain.PageQuery) ([]Event, int64, error)
	ListOpenByDeviceSince(context.Context, string, string, string, time.Time) ([]Event, error)
}
