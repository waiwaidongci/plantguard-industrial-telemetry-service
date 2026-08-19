package domain

import (
	"context"
	"time"

	shareddomain "github.com/acme/plantguard/internal/shared/domain"
)

type Device struct {
	ID              string    `json:"id"`
	TenantID        string    `json:"tenant_id"`
	SiteID          string    `json:"site_id"`
	ModelID         string    `json:"model_id"`
	Name            string    `json:"name"`
	Protocol        string    `json:"protocol"`
	Enabled         bool      `json:"enabled"`
	FirmwareVersion string    `json:"firmware_version"`
	Tags            []string  `json:"tags"`
	LastSeenAt      time.Time `json:"last_seen_at"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	Version         int64     `json:"version"`
}

func NewDevice(tenantID, siteID, modelID, name, protocol, firmware string, tags []string, enabled bool, now time.Time) Device {
	if protocol == "" {
		protocol = "mqtt"
	}
	return Device{
		ID:              shareddomain.NewID("dev"),
		TenantID:        tenantID,
		SiteID:          siteID,
		ModelID:         modelID,
		Name:            name,
		Protocol:        protocol,
		Enabled:         enabled,
		FirmwareVersion: firmware,
		Tags:            tags,
		Status:          "unknown",
		CreatedAt:       now,
		UpdatedAt:       now,
		Version:         1,
	}
}

type DeviceRepository interface {
	Create(context.Context, Device) error
	GetByID(context.Context, string, string) (Device, error)
	Update(context.Context, Device) error
	List(context.Context, string, shareddomain.PageQuery) ([]Device, int64, error)
	UpdateHeartbeat(context.Context, string, string, time.Time) error
	ListByTenant(context.Context, string) ([]Device, error)
}
