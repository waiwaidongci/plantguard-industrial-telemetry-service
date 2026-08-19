package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"

	sharedinfra "github.com/acme/plantguard/internal/shared/infrastructure"
	telemetrydomain "github.com/acme/plantguard/internal/telemetry/domain"
)

type DeviceReader struct {
	db *sql.DB
}

func NewDeviceReader(db *sql.DB) *DeviceReader {
	return &DeviceReader{db: db}
}

func (r *DeviceReader) GetSnapshot(ctx context.Context, tenantID, deviceID string) (telemetrydomain.DeviceSnapshot, error) {
	var snapshot telemetrydomain.DeviceSnapshot
	var modelID sql.NullString
	var enabled int
	err := r.db.QueryRowContext(ctx, `SELECT id, tenant_id, model_id, enabled FROM devices WHERE tenant_id=? AND id=?`, tenantID, deviceID).
		Scan(&snapshot.ID, &snapshot.TenantID, &modelID, &enabled)
	if err != nil {
		return snapshot, sharedinfra.MapSQLError(err, nil)
	}
	snapshot.ModelID = modelID.String
	snapshot.Enabled = enabled != 0
	return snapshot, nil
}

type ModelMetricReader struct {
	db *sql.DB
}

func NewModelMetricReader(db *sql.DB) *ModelMetricReader {
	return &ModelMetricReader{db: db}
}

func (r *ModelMetricReader) GetMetricSpecs(ctx context.Context, tenantID, modelID string) (map[string]telemetrydomain.MetricSpec, error) {
	specs := map[string]telemetrydomain.MetricSpec{}
	if modelID == "" {
		return specs, nil
	}
	var raw string
	err := r.db.QueryRowContext(ctx, `SELECT metric_specs FROM device_models WHERE tenant_id=? AND id=?`, tenantID, modelID).Scan(&raw)
	if err != nil {
		return specs, sharedinfra.MapSQLError(err, nil)
	}
	var decoded []telemetrydomain.MetricSpec
	if err := json.Unmarshal([]byte(raw), &decoded); err != nil {
		return specs, err
	}
	for _, spec := range decoded {
		specs[spec.Key] = spec
	}
	return specs, nil
}
