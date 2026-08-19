package application

import (
	"context"
	"errors"
	"testing"
	"time"

	shareddomain "github.com/acme/plantguard/internal/shared/domain"
	telemetrydomain "github.com/acme/plantguard/internal/telemetry/domain"
)

type okDeviceReader struct{}

func (okDeviceReader) GetSnapshot(context.Context, string, string) (telemetrydomain.DeviceSnapshot, error) {
	return telemetrydomain.DeviceSnapshot{ID: "dev-1", TenantID: "tenant-1", ModelID: "model-1", Enabled: true}, nil
}

type okModelReader struct{}

func (okModelReader) GetMetricSpecs(context.Context, string, string) (map[string]telemetrydomain.MetricSpec, error) {
	return map[string]telemetrydomain.MetricSpec{
		"pressure": {Key: "pressure"},
	}, nil
}

type capturingRepo struct {
	got        context.Context
	summaryCtx context.Context
}

func (r *capturingRepo) StoreBatch(ctx context.Context, _ telemetrydomain.Batch) error {
	r.got = ctx
	return ctx.Err()
}
func (r *capturingRepo) Summaries(ctx context.Context, _, _ string, _, _ time.Time) ([]telemetrydomain.Summary, error) {
	r.summaryCtx = ctx
	return nil, ctx.Err()
}
func (capturingRepo) ListReadings(context.Context, string, string, shareddomain.PageQuery) ([]telemetrydomain.ReadingRecord, int64, error) {
	return nil, 0, nil
}
func (capturingRepo) CountSince(context.Context, string, string, string, time.Time) (int, error) { return 0, nil }
func (capturingRepo) DeleteSummariesBefore(context.Context, time.Time) (int64, error)          { return 0, nil }

type clockNow struct{}

func (clockNow) Now(context.Context) time.Time { return time.Unix(1700000000, 0).UTC() }

func TestIngestBatchHonorsCanceledContext(t *testing.T) {
	repo := &capturingRepo{}
	svc := NewService(okDeviceReader{}, okModelReader{}, repo, clockNow{})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	t.Run("ingest", func(t *testing.T) {
		_, err := svc.IngestBatch(ctx, "tenant-1", "dev-1", IngestBatchInput{
			Source: "test",
			Readings: []telemetrydomain.Reading{
				{Metric: "pressure", Value: 1, Timestamp: time.Unix(1700000000, 0).UTC()},
			},
		})
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context.Canceled, got %v", err)
		}
		if repo.got == nil || repo.got.Err() != context.Canceled {
			t.Fatalf("repository did not receive canceled context: %#v", repo.got)
		}
	})
	t.Run("summaries", func(t *testing.T) {
		_, err := svc.Summaries(ctx, "tenant-1", "dev-1", SummaryQuery{})
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context.Canceled, got %v", err)
		}
		if repo.summaryCtx == nil || repo.summaryCtx.Err() != context.Canceled {
			t.Fatalf("repository did not receive canceled context for summaries: %#v", repo.summaryCtx)
		}
	})
}
