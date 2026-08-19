package domain

import (
	"context"
	"encoding/json"
	"time"

	shareddomain "github.com/acme/plantguard/internal/shared/domain"
)

type MetricType string

const (
	MetricTemperature MetricType = "temperature"
	MetricPressure    MetricType = "pressure"
	MetricVoltage     MetricType = "voltage"
	MetricRPM         MetricType = "rpm"
	MetricSwitch      MetricType = "switch"
)

type MetricSpec struct {
	Key       string     `json:"key"`
	Type      MetricType `json:"type"`
	Unit      string     `json:"unit"`
	Min       *float64   `json:"min,omitempty"`
	Max       *float64   `json:"max,omitempty"`
	Precision int        `json:"precision"`
	Required  bool       `json:"required"`
}

type DeviceModel struct {
	ID          string       `json:"id"`
	TenantID    string       `json:"tenant_id"`
	Name        string       `json:"name"`
	MetricSpecs []MetricSpec `json:"metric_specs"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	Version     int64        `json:"version"`
}

func NewDeviceModel(tenantID, name string, specs []MetricSpec, now time.Time) DeviceModel {
	return DeviceModel{
		ID:          shareddomain.NewID("mdl"),
		TenantID:    tenantID,
		Name:        name,
		MetricSpecs: specs,
		CreatedAt:   now,
		UpdatedAt:   now,
		Version:     1,
	}
}

func (m DeviceModel) SpecMap() map[string]MetricSpec {
	out := make(map[string]MetricSpec, len(m.MetricSpecs))
	for _, spec := range m.MetricSpecs {
		out[spec.Key] = spec
	}
	return out
}

func (m DeviceModel) MarshalSpecs() (string, error) {
	raw, err := json.Marshal(m.MetricSpecs)
	return string(raw), err
}

func UnmarshalSpecs(raw string) ([]MetricSpec, error) {
	var specs []MetricSpec
	if err := json.Unmarshal([]byte(raw), &specs); err != nil {
		return nil, err
	}
	return specs, nil
}

type DeviceModelRepository interface {
	Create(context.Context, DeviceModel) error
	GetByID(context.Context, string, string) (DeviceModel, error)
	Update(context.Context, DeviceModel) error
	List(context.Context, string, shareddomain.PageQuery) ([]DeviceModel, int64, error)
}
