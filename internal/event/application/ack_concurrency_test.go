package application

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	eventdomain "github.com/acme/plantguard/internal/event/domain"
	shareddomain "github.com/acme/plantguard/internal/shared/domain"
)

type racingEventRepo struct{}

func (racingEventRepo) Create(context.Context, eventdomain.Event) error { return nil }
func (racingEventRepo) GetByID(context.Context, string, string) (eventdomain.Event, error) {
	return eventdomain.Event{
		ID:        "event-1",
		TenantID:  "tenant-1",
		DeviceID:  "device-1",
		Status:    "open",
		Version:   1,
		CreatedAt: time.Unix(1700000000, 0).UTC(),
		UpdatedAt: time.Unix(1700000000, 0).UTC(),
	}, nil
}
func (racingEventRepo) Update(context.Context, eventdomain.Event) error { return nil }
func (racingEventRepo) List(context.Context, string, shareddomain.PageQuery) ([]eventdomain.Event, int64, error) {
	return nil, 0, nil
}
func (racingEventRepo) ListOpenByDeviceSince(context.Context, string, string, string, time.Time) ([]eventdomain.Event, error) {
	return nil, nil
}

type ackClock struct{}

func (ackClock) Now(context.Context) time.Time { return time.Unix(1700000000, 0).UTC() }

func TestConcurrentAcknowledgeOnlyOnce(t *testing.T) {
	svc := NewService(racingEventRepo{}, nil, ackClock{})
	start := make(chan struct{})
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	wg.Add(2)
	for i := 0; i < 2; i++ {
		go func() {
			defer wg.Done()
			<-start
			_, err := svc.Acknowledge(context.Background(), "tenant-1", "event-1", AcknowledgeEventInput{AcknowledgedBy: "operator"})
			errs <- err
		}()
	}
	close(start)
	wg.Wait()
	close(errs)

	success := 0
	already := 0
	for err := range errs {
		switch {
		case err == nil:
			success++
		case errors.Is(err, eventdomain.ErrAlreadyAcknowledged):
			already++
		default:
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if success != 1 || already != 1 {
		t.Fatalf("expected one success and one already-acknowledged, got success=%d already=%d", success, already)
	}
}

func TestNormalizeEventSeverity(t *testing.T) {
	if got := normalizeSeverity("critical"); got != "critical" {
		t.Fatalf("expected critical severity, got %q", got)
	}
}
