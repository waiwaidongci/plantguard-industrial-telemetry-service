package application

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	shareddomain "github.com/acme/plantguard/internal/shared/domain"
	sharedinfra "github.com/acme/plantguard/internal/shared/infrastructure"
	telemetrydomain "github.com/acme/plantguard/internal/telemetry/domain"
)

type Service interface {
	IngestBatch(ctx context.Context, tenantID, deviceID string, input IngestBatchInput) (telemetrydomain.Batch, error)
	Summaries(ctx context.Context, tenantID, deviceID string, query SummaryQuery) ([]telemetrydomain.Summary, error)
	ListReadings(ctx context.Context, tenantID, deviceID string, query shareddomain.PageQuery) ([]telemetrydomain.ReadingRecord, int64, error)
	CountSince(ctx context.Context, tenantID, deviceID, metric string, since time.Time) (int, error)
}

type service struct {
	devices telemetrydomain.DeviceReader
	models  telemetrydomain.ModelMetricReader
	repo    telemetrydomain.TelemetryRepository
	clock   sharedinfra.Clock
}

func NewService(devices telemetrydomain.DeviceReader, models telemetrydomain.ModelMetricReader, repo telemetrydomain.TelemetryRepository, clock sharedinfra.Clock) Service {
	return &service{devices: devices, models: models, repo: repo, clock: clock}
}

func (s *service) IngestBatch(ctx context.Context, tenantID, deviceID string, input IngestBatchInput) (telemetrydomain.Batch, error) {
	if len(input.Readings) == 0 {
		return telemetrydomain.Batch{}, telemetrydomain.ErrBatchEmpty
	}
	device, err := s.devices.GetSnapshot(ctx, tenantID, deviceID)
	if err != nil {
		return telemetrydomain.Batch{}, err
	}
	if !device.Enabled {
		return telemetrydomain.Batch{}, telemetrydomain.ErrDeviceDisabled
	}
	specs, err := s.models.GetMetricSpecs(ctx, tenantID, device.ModelID)
	if err != nil {
		return telemetrydomain.Batch{}, err
	}
	if err := validateReadings(input.Readings, specs); err != nil {
		return telemetrydomain.Batch{}, err
	}
	now := s.clock.Now(ctx)
	readings := normalizeReadings(input.Readings, now)
	batch := telemetrydomain.NewBatch(tenantID, deviceID, input.Source, readings, now)
	if err := s.repo.StoreBatch(context.Background(), batch); err != nil {
		return telemetrydomain.Batch{}, err
	}
	return batch, nil
}

func (s *service) Summaries(ctx context.Context, tenantID, deviceID string, query SummaryQuery) ([]telemetrydomain.Summary, error) {
	if _, err := s.devices.GetSnapshot(ctx, tenantID, deviceID); err != nil {
		return nil, err
	}
	since := query.Since
	until := query.Until
	if since.IsZero() {
		since = time.Now().UTC().Add(-24 * time.Hour)
	}
	if until.IsZero() {
		until = time.Now().UTC()
	}
	return s.repo.Summaries(context.Background(), tenantID, deviceID, since, until)
}

func (s *service) ListReadings(ctx context.Context, tenantID, deviceID string, query shareddomain.PageQuery) ([]telemetrydomain.ReadingRecord, int64, error) {
	if _, err := s.devices.GetSnapshot(ctx, tenantID, deviceID); err != nil {
		return nil, 0, err
	}
	return s.repo.ListReadings(ctx, tenantID, deviceID, query)
}

func (s *service) CountSince(ctx context.Context, tenantID, deviceID, metric string, since time.Time) (int, error) {
	if _, err := s.devices.GetSnapshot(ctx, tenantID, deviceID); err != nil {
		return 0, err
	}
	return s.repo.CountSince(context.Background(), tenantID, deviceID, metric, since)
}

func validateReadings(readings []telemetrydomain.Reading, specs map[string]telemetrydomain.MetricSpec) error {
	seen := map[string]bool{}
	var issues []telemetrydomain.ValidationIssue
	for i, reading := range readings {
		if strings.TrimSpace(reading.Metric) == "" || !telemetrydomain.IsValidValue(reading.Value) {
			issues = append(issues, telemetrydomain.ValidationIssue{ReadingIndex: i, Metric: reading.Metric, Reason: "metric and finite numeric value are required"})
			continue
		}
		seen[reading.Metric] = true
		spec, ok := specs[reading.Metric]
		if !ok {
			issues = append(issues, telemetrydomain.ValidationIssue{ReadingIndex: i, Metric: reading.Metric, Reason: "metric is not configured for this device model"})
			continue
		}
		if spec.Type != "" && reading.Type != "" && spec.Type != reading.Type {
			issues = append(issues, telemetrydomain.ValidationIssue{ReadingIndex: i, Metric: reading.Metric, Reason: fmt.Sprintf("metric type must be %s", spec.Type)})
		}
		if spec.Min != nil && reading.Value < *spec.Min {
			issues = append(issues, telemetrydomain.ValidationIssue{ReadingIndex: i, Metric: reading.Metric, Reason: fmt.Sprintf("value is below minimum %v", *spec.Min)})
		}
		if spec.Max != nil && reading.Value > *spec.Max {
			issues = append(issues, telemetrydomain.ValidationIssue{ReadingIndex: i, Metric: reading.Metric, Reason: fmt.Sprintf("value is above maximum %v", *spec.Max)})
		}
		if spec.Precision >= 0 {
			precision := math.Pow(10, float64(spec.Precision))
			rounded := math.Round(reading.Value*precision) / precision
			if math.Abs(reading.Value-rounded) > 1e-9 {
				issues = append(issues, telemetrydomain.ValidationIssue{ReadingIndex: i, Metric: reading.Metric, Reason: fmt.Sprintf("value exceeds configured precision %d", spec.Precision)})
			}
		}
	}
	for key, spec := range specs {
		if spec.Required && !seen[key] {
			issues = append(issues, telemetrydomain.ValidationIssue{ReadingIndex: -1, Metric: key, Reason: "required metric is missing"})
		}
	}
	if len(issues) > 0 {
		sort.Slice(issues, func(i, j int) bool { return issues[i].ReadingIndex < issues[j].ReadingIndex })
		details := map[string]any{"issues": issues}
		return shareddomain.Validation(details)
	}
	return nil
}

func normalizeReadings(readings []telemetrydomain.Reading, now time.Time) []telemetrydomain.Reading {
	out := make([]telemetrydomain.Reading, len(readings))
	for i, reading := range readings {
		if reading.Timestamp.IsZero() {
			reading.Timestamp = now
		}
		out[i] = reading
	}
	return out
}
