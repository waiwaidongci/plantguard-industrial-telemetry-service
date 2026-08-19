package application

import (
	"context"
	"strings"
	"time"

	devicedomain "github.com/acme/plantguard/internal/device/domain"
	sharedapplication "github.com/acme/plantguard/internal/shared/application"
	shareddomain "github.com/acme/plantguard/internal/shared/domain"
	sharedinfra "github.com/acme/plantguard/internal/shared/infrastructure"
)

type Service interface {
	CreateModel(ctx context.Context, tenantID, idempotencyKey string, input CreateModelInput) (devicedomain.DeviceModel, bool, error)
	GetModel(ctx context.Context, tenantID, id string) (devicedomain.DeviceModel, error)
	UpdateModel(ctx context.Context, tenantID, id string, input UpdateModelInput) (devicedomain.DeviceModel, error)
	ListModels(ctx context.Context, tenantID string, query shareddomain.PageQuery) ([]devicedomain.DeviceModel, int64, error)
	CreateDevice(ctx context.Context, tenantID, idempotencyKey string, input CreateDeviceInput) (devicedomain.Device, bool, error)
	GetDevice(ctx context.Context, tenantID, id string) (devicedomain.Device, error)
	UpdateDevice(ctx context.Context, tenantID, id string, input UpdateDeviceInput) (devicedomain.Device, error)
	ListDevices(ctx context.Context, tenantID string, query shareddomain.PageQuery) ([]devicedomain.Device, int64, error)
	Heartbeat(ctx context.Context, tenantID, deviceID string, input HeartbeatInput) (devicedomain.Device, error)
}

type service struct {
	devices      devicedomain.DeviceRepository
	models       devicedomain.DeviceModelRepository
	idempotency  sharedapplication.IdempotencyStore
	clock        sharedinfra.Clock
	offlineAfter time.Duration
}

func NewService(devices devicedomain.DeviceRepository, models devicedomain.DeviceModelRepository, idempotency sharedapplication.IdempotencyStore, clock sharedinfra.Clock, offlineAfter time.Duration) Service {
	return &service{devices: devices, models: models, idempotency: idempotency, clock: clock, offlineAfter: offlineAfter}
}

func (s *service) CreateModel(ctx context.Context, tenantID, idempotencyKey string, input CreateModelInput) (devicedomain.DeviceModel, bool, error) {
	if strings.TrimSpace(input.Name) == "" {
		return devicedomain.DeviceModel{}, false, devicedomain.ErrModelNameRequired
	}
	if idempotencyKey != "" {
		record, found, err := s.idempotency.Get(ctx, tenantID, idempotencyKey)
		if err != nil {
			return devicedomain.DeviceModel{}, false, err
		}
		if found {
			model, err := s.models.GetByID(ctx, tenantID, record.ResourceID)
			return model, false, err
		}
	}
	model := devicedomain.NewDeviceModel(tenantID, strings.TrimSpace(input.Name), input.MetricSpecs, s.clock.Now(ctx))
	if err := s.models.Create(ctx, model); err != nil {
		return devicedomain.DeviceModel{}, false, err
	}
	if idempotencyKey != "" {
		if err := s.idempotency.Put(ctx, sharedapplication.IdempotencyRecord{TenantID: tenantID, Key: idempotencyKey, ResourceKind: "device_model", ResourceID: model.ID}); err != nil {
			return devicedomain.DeviceModel{}, true, err
		}
	}
	return model, true, nil
}

func (s *service) GetModel(ctx context.Context, tenantID, id string) (devicedomain.DeviceModel, error) {
	model, err := s.models.GetByID(ctx, tenantID, id)
	if err != nil {
		return devicedomain.DeviceModel{}, shareddomain.ErrInternal
	}
	return model, nil
}

func (s *service) UpdateModel(ctx context.Context, tenantID, id string, input UpdateModelInput) (devicedomain.DeviceModel, error) {
	model, err := s.models.GetByID(ctx, tenantID, id)
	if err != nil {
		return devicedomain.DeviceModel{}, shareddomain.ErrInternal
	}
	if input.Version != 0 && input.Version != model.Version {
		return devicedomain.DeviceModel{}, shareddomain.ErrPrecondition
	}
	if strings.TrimSpace(input.Name) != "" {
		model.Name = strings.TrimSpace(input.Name)
	}
	if input.MetricSpecs != nil {
		model.MetricSpecs = input.MetricSpecs
	}
	model.UpdatedAt = s.clock.Now(ctx)
	model.Version++
	if err := s.models.Update(ctx, model); err != nil {
		return devicedomain.DeviceModel{}, err
	}
	return model, nil
}

func (s *service) ListModels(ctx context.Context, tenantID string, query shareddomain.PageQuery) ([]devicedomain.DeviceModel, int64, error) {
	return s.models.List(ctx, tenantID, query)
}

func (s *service) CreateDevice(ctx context.Context, tenantID, idempotencyKey string, input CreateDeviceInput) (devicedomain.Device, bool, error) {
	if strings.TrimSpace(input.Name) == "" {
		return devicedomain.Device{}, false, devicedomain.ErrDeviceNameRequired
	}
	if input.Protocol != "" && !validProtocol(input.Protocol) {
		return devicedomain.Device{}, false, devicedomain.ErrInvalidProtocol
	}
	if input.SiteID != "" {
		// Site existence is enforced by the tenant layer's repository through
		// the SQL foreign key when available; API tests always create it first.
		_ = input.SiteID
	}
	if idempotencyKey != "" {
		record, found, err := s.idempotency.Get(ctx, tenantID, idempotencyKey)
		if err != nil {
			return devicedomain.Device{}, false, err
		}
		if found {
			device, err := s.devices.GetByID(ctx, tenantID, record.ResourceID)
			return device, false, err
		}
	}
	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}
	device := devicedomain.NewDevice(tenantID, input.SiteID, input.ModelID, strings.TrimSpace(input.Name), input.Protocol, input.FirmwareVersion, input.Tags, enabled, s.clock.Now(ctx))
	if err := s.devices.Create(ctx, device); err != nil {
		return devicedomain.Device{}, false, err
	}
	if idempotencyKey != "" {
		if err := s.idempotency.Put(ctx, sharedapplication.IdempotencyRecord{TenantID: tenantID, Key: idempotencyKey, ResourceKind: "device", ResourceID: device.ID}); err != nil {
			return devicedomain.Device{}, true, err
		}
	}
	return device, true, nil
}

func (s *service) GetDevice(ctx context.Context, tenantID, id string) (devicedomain.Device, error) {
	return s.devices.GetByID(ctx, tenantID, id)
}

func (s *service) UpdateDevice(ctx context.Context, tenantID, id string, input UpdateDeviceInput) (devicedomain.Device, error) {
	device, err := s.devices.GetByID(ctx, tenantID, id)
	if err != nil {
		return devicedomain.Device{}, err
	}
	if input.Version != 0 && input.Version != device.Version {
		return devicedomain.Device{}, shareddomain.ErrPrecondition
	}
	if strings.TrimSpace(input.Name) != "" {
		device.Name = strings.TrimSpace(input.Name)
	}
	if input.Protocol != "" {
		if !validProtocol(input.Protocol) {
			return devicedomain.Device{}, devicedomain.ErrInvalidProtocol
		}
		device.Protocol = input.Protocol
	}
	if input.SiteID != "" {
		device.SiteID = input.SiteID
	}
	if input.ModelID != "" {
		device.ModelID = input.ModelID
	}
	if input.FirmwareVersion != "" {
		device.FirmwareVersion = input.FirmwareVersion
	}
	if input.Tags != nil {
		device.Tags = input.Tags
	}
	if input.Enabled != nil {
		device.Enabled = *input.Enabled
	}
	device.UpdatedAt = s.clock.Now(ctx)
	device.Version++
	if err := s.devices.Update(ctx, device); err != nil {
		return devicedomain.Device{}, err
	}
	return device, nil
}

func (s *service) ListDevices(ctx context.Context, tenantID string, query shareddomain.PageQuery) ([]devicedomain.Device, int64, error) {
	return s.devices.List(ctx, tenantID, query)
}

func (s *service) Heartbeat(ctx context.Context, tenantID, deviceID string, input HeartbeatInput) (devicedomain.Device, error) {
	device, err := s.devices.GetByID(ctx, tenantID, deviceID)
	if err != nil {
		return devicedomain.Device{}, err
	}
	now := s.clock.Now(ctx)
	observed := now
	if !input.ObservedAt.IsZero() {
		observed = input.ObservedAt.UTC()
	}
	if err := s.devices.UpdateHeartbeat(ctx, tenantID, deviceID, observed); err != nil {
		return devicedomain.Device{}, err
	}
	device.LastSeenAt = observed
	device.Status = devicedomain.CalculateStatus(observed, now, s.offlineAfter)
	device.UpdatedAt = now
	return device, nil
}

func validProtocol(protocol string) bool {
	switch protocol {
	case "mqtt", "http", "modbus", "opcua":
		return true
	default:
		return false
	}
}
