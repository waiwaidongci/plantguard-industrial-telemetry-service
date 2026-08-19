package application

import (
	"context"
	"errors"
	"testing"
	"time"

	devicedomain "github.com/acme/plantguard/internal/device/domain"
	shareddomain "github.com/acme/plantguard/internal/shared/domain"
)

type missingModelRepo struct{}

func (missingModelRepo) Create(context.Context, devicedomain.DeviceModel) error { return nil }
func (missingModelRepo) GetByID(context.Context, string, string) (devicedomain.DeviceModel, error) {
	return devicedomain.DeviceModel{}, shareddomain.ErrNotFound
}
func (missingModelRepo) Update(context.Context, devicedomain.DeviceModel) error { return nil }
func (missingModelRepo) List(context.Context, string, shareddomain.PageQuery) ([]devicedomain.DeviceModel, int64, error) {
	return nil, 0, nil
}

type fixedClock struct{}

func (fixedClock) Now(context.Context) time.Time { return time.Unix(1700000000, 0).UTC() }

func TestGetModelPreservesNotFoundError(t *testing.T) {
	t.Run("get missing model", func(t *testing.T) {
		svc := NewService(nil, missingModelRepo{}, nil, fixedClock{}, time.Minute)
		_, err := svc.GetModel(context.Background(), "tenant-1", "missing-model")
		if !errors.Is(err, shareddomain.ErrNotFound) {
			t.Fatalf("expected not-found error to be preserved, got %T: %v", err, err)
		}
	})
	t.Run("update missing model", func(t *testing.T) {
		svc := NewService(nil, missingModelRepo{}, nil, fixedClock{}, time.Minute)
		_, err := svc.UpdateModel(context.Background(), "tenant-1", "missing-model", UpdateModelInput{Name: "new"})
		if !errors.Is(err, shareddomain.ErrNotFound) {
			t.Fatalf("expected not-found error to be preserved, got %T: %v", err, err)
		}
	})
}
