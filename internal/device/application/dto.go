package application

import (
	"time"

	devicedomain "github.com/acme/plantguard/internal/device/domain"
)

type CreateModelInput struct {
	Name        string                    `json:"name"`
	MetricSpecs []devicedomain.MetricSpec `json:"metric_specs"`
}

type UpdateModelInput struct {
	Name        string                    `json:"name"`
	MetricSpecs []devicedomain.MetricSpec `json:"metric_specs"`
	Version     int64                     `json:"version"`
}

type CreateDeviceInput struct {
	SiteID          string   `json:"site_id"`
	ModelID         string   `json:"model_id"`
	Name            string   `json:"name"`
	Protocol        string   `json:"protocol"`
	Enabled         *bool    `json:"enabled,omitempty"`
	FirmwareVersion string   `json:"firmware_version"`
	Tags            []string `json:"tags"`
}

type UpdateDeviceInput struct {
	SiteID          string   `json:"site_id"`
	ModelID         string   `json:"model_id"`
	Name            string   `json:"name"`
	Protocol        string   `json:"protocol"`
	Enabled         *bool    `json:"enabled,omitempty"`
	FirmwareVersion string   `json:"firmware_version"`
	Tags            []string `json:"tags"`
	Version         int64    `json:"version"`
}

type HeartbeatInput struct {
	ObservedAt time.Time `json:"observed_at"`
}
