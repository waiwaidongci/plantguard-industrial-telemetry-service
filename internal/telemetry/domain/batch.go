package domain

import (
	"context"
	"math"
	"time"

	shareddomain "github.com/acme/plantguard/internal/shared/domain"
)

type Reading struct {
	Metric    string    `json:"metric"`
	Value     float64   `json:"value"`
	Unit      string    `json:"unit"`
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
}

type Batch struct {
	ID         string    `json:"id"`
	TenantID   string    `json:"tenant_id"`
	DeviceID   string    `json:"device_id"`
	Source     string    `json:"source"`
	ReceivedAt time.Time `json:"received_at"`
	Readings   []Reading `json:"readings"`
}

func NewBatch(tenantID, deviceID, source string, readings []Reading, now time.Time) Batch {
	return Batch{
		ID:         shareddomain.NewID("tel"),
		TenantID:   tenantID,
		DeviceID:   deviceID,
		Source:     source,
		ReceivedAt: now,
		Readings:   readings,
	}
}

func (b Batch) Valid() bool {
	return len(b.Readings) > 0 && b.TenantID != "" && b.DeviceID != ""
}

type ValidationIssue struct {
	ReadingIndex int    `json:"reading_index"`
	Metric       string `json:"metric"`
	Reason       string `json:"reason"`
}

type DeviceSnapshot struct {
	ID       string
	TenantID string
	ModelID  string
	Enabled  bool
}

type DeviceReader interface {
	GetSnapshot(context.Context, string, string) (DeviceSnapshot, error)
}

type MetricSpec struct {
	Key       string
	Type      string
	Unit      string
	Min       *float64
	Max       *float64
	Precision int
	Required  bool
}

type ModelMetricReader interface {
	GetMetricSpecs(context.Context, string, string) (map[string]MetricSpec, error)
}

type TelemetryRepository interface {
	StoreBatch(context.Context, Batch) error
	Summaries(context.Context, string, string, time.Time, time.Time) ([]Summary, error)
	ListReadings(context.Context, string, string, shareddomain.PageQuery) ([]ReadingRecord, int64, error)
	CountSince(context.Context, string, string, string, time.Time) (int, error)
	DeleteSummariesBefore(context.Context, time.Time) (int64, error)
}

type Summary struct {
	DeviceID    string    `json:"device_id"`
	Metric      string    `json:"metric"`
	WindowStart time.Time `json:"window_start"`
	WindowEnd   time.Time `json:"window_end"`
	SampleCount int       `json:"sample_count"`
	Min         float64   `json:"min"`
	Max         float64   `json:"max"`
	Avg         float64   `json:"avg"`
	Last        float64   `json:"last"`
}

type ReadingRecord struct {
	DeviceID   string    `json:"device_id"`
	Metric     string    `json:"metric"`
	Value      float64   `json:"value"`
	Unit       string    `json:"unit"`
	ReceivedAt time.Time `json:"received_at"`
}

func IsValidValue(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
