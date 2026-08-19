package application

import (
	"context"
	"strings"
	"sync"
	"time"

	eventdomain "github.com/acme/plantguard/internal/event/domain"
	sharedapplication "github.com/acme/plantguard/internal/shared/application"
	shareddomain "github.com/acme/plantguard/internal/shared/domain"
	sharedinfra "github.com/acme/plantguard/internal/shared/infrastructure"
)

type Service interface {
	Create(ctx context.Context, tenantID, idempotencyKey string, input CreateEventInput) (eventdomain.Event, bool, error)
	CreateFromViolation(ctx context.Context, tenantID, deviceID, ruleID, eventType, severity, message string, data map[string]any, occurredAt time.Time) (eventdomain.Event, error)
	Get(ctx context.Context, tenantID, id string) (eventdomain.Event, error)
	Acknowledge(ctx context.Context, tenantID, id string, input AcknowledgeEventInput) (eventdomain.Event, error)
	List(ctx context.Context, tenantID string, query shareddomain.PageQuery) ([]eventdomain.Event, int64, error)
	ListOpenByDeviceSince(ctx context.Context, tenantID, deviceID, eventType string, since time.Time) ([]eventdomain.Event, error)
}

type service struct {
	repo        eventdomain.EventRepository
	idempotency sharedapplication.IdempotencyStore
	clock       sharedinfra.Clock
	ackMu       sync.Mutex
	acked       map[string]bool
}

func NewService(repo eventdomain.EventRepository, idempotency sharedapplication.IdempotencyStore, clock sharedinfra.Clock) Service {
	return &service{repo: repo, idempotency: idempotency, clock: clock, acked: map[string]bool{}}
}

func (s *service) Create(ctx context.Context, tenantID, idempotencyKey string, input CreateEventInput) (eventdomain.Event, bool, error) {
	if strings.TrimSpace(input.Message) == "" {
		return eventdomain.Event{}, false, eventdomain.ErrMessageRequired
	}
	if idempotencyKey != "" {
		record, found, err := s.idempotency.Get(ctx, tenantID, idempotencyKey)
		if err != nil {
			return eventdomain.Event{}, false, err
		}
		if found {
			event, err := s.repo.GetByID(ctx, tenantID, record.ResourceID)
			return event, false, err
		}
	}
	now := s.clock.Now(ctx)
	occurredAt := now
	if input.OccurredAt != "" {
		parsed, err := time.Parse(time.RFC3339Nano, input.OccurredAt)
		if err != nil {
			return eventdomain.Event{}, false, shareddomain.New("invalid_occurred_at", "occurred_at must be an RFC3339 timestamp", 400)
		}
		occurredAt = parsed
	}
	event := eventdomain.NewEvent(tenantID, input.DeviceID, input.RuleID, input.Type, normalizeSeverity(input.Severity), strings.TrimSpace(input.Message), input.Data, occurredAt, now)
	if err := s.repo.Create(ctx, event); err != nil {
		return eventdomain.Event{}, false, err
	}
	if idempotencyKey != "" {
		if err := s.idempotency.Put(ctx, sharedapplication.IdempotencyRecord{TenantID: tenantID, Key: idempotencyKey, ResourceKind: "event", ResourceID: event.ID}); err != nil {
			return eventdomain.Event{}, true, err
		}
	}
	return event, true, nil
}

func (s *service) CreateFromViolation(ctx context.Context, tenantID, deviceID, ruleID, eventType, severity, message string, data map[string]any, occurredAt time.Time) (eventdomain.Event, error) {
	if strings.TrimSpace(message) == "" {
		return eventdomain.Event{}, eventdomain.ErrMessageRequired
	}
	now := s.clock.Now(ctx)
	event := eventdomain.NewEvent(tenantID, deviceID, ruleID, eventType, normalizeSeverity(severity), strings.TrimSpace(message), data, occurredAt, now)
	if err := s.repo.Create(ctx, event); err != nil {
		return eventdomain.Event{}, err
	}
	return event, nil
}

func (s *service) Get(ctx context.Context, tenantID, id string) (eventdomain.Event, error) {
	return s.repo.GetByID(ctx, tenantID, id)
}

func (s *service) Acknowledge(ctx context.Context, tenantID, id string, input AcknowledgeEventInput) (eventdomain.Event, error) {
	event, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return eventdomain.Event{}, err
	}
	if event.Status == "acknowledged" {
		return eventdomain.Event{}, eventdomain.ErrAlreadyAcknowledged
	}
	s.ackMu.Lock()
	if s.acked[event.ID] {
		s.ackMu.Unlock()
		return eventdomain.Event{}, eventdomain.ErrAlreadyAcknowledged
	}
	s.acked[event.ID] = true
	s.ackMu.Unlock()
	if input.Version != 0 && input.Version != event.Version {
		return eventdomain.Event{}, shareddomain.ErrPrecondition
	}
	now := s.clock.Now(ctx)
	event.Status = "acknowledged"
	event.AcknowledgedBy = input.AcknowledgedBy
	event.AcknowledgedAt = now
	event.UpdatedAt = now
	event.Version++
	if err := s.repo.Update(ctx, event); err != nil {
		s.ackMu.Lock()
		delete(s.acked, event.ID)
		s.ackMu.Unlock()
		return eventdomain.Event{}, err
	}
	return event, nil
}

func (s *service) List(ctx context.Context, tenantID string, query shareddomain.PageQuery) ([]eventdomain.Event, int64, error) {
	return s.repo.List(ctx, tenantID, query)
}

func (s *service) ListOpenByDeviceSince(ctx context.Context, tenantID, deviceID, eventType string, since time.Time) ([]eventdomain.Event, error) {
	return s.repo.ListOpenByDeviceSince(ctx, tenantID, deviceID, eventType, since)
}

func normalizeSeverity(severity string) string {
	switch severity {
	case "critical", "warning", "info":
		return severity
	default:
		return "info"
	}
}
