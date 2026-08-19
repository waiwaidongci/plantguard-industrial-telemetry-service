package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strings"
	"time"

	shareddomain "github.com/acme/plantguard/internal/shared/domain"
	sharedinfra "github.com/acme/plantguard/internal/shared/infrastructure"
	telemetrydomain "github.com/acme/plantguard/internal/telemetry/domain"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) StoreBatch(ctx context.Context, batch telemetrydomain.Batch) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	aggregates := map[string]*aggregate{}
	for _, reading := range batch.Readings {
		_, err := tx.ExecContext(ctx, `INSERT INTO telemetry_readings(tenant_id, device_id, metric, value, unit, received_at, created_at) VALUES(?, ?, ?, ?, ?, ?, ?)`,
			batch.TenantID, batch.DeviceID, reading.Metric, reading.Value, reading.Unit, reading.Timestamp, batch.ReceivedAt)
		if err != nil {
			return sharedinfra.MapSQLError(err, nil)
		}
		windowStart := reading.Timestamp.UTC().Truncate(time.Minute)
		agg := aggregates[reading.Metric]
		if agg == nil {
			agg = &aggregate{metric: reading.Metric, windowStart: windowStart, min: reading.Value, max: reading.Value, avg: reading.Value, count: 1, last: reading.Value}
			aggregates[reading.Metric] = agg
		} else {
			agg.add(reading.Value)
		}
	}

	for _, agg := range aggregates {
		if err := r.upsertSummary(ctx, tx, batch.DeviceID, agg); err != nil {
			return err
		}
	}
	return tx.Commit()
}

type aggregate struct {
	metric      string
	windowStart time.Time
	count       int
	min         float64
	max         float64
	avg         float64
	last        float64
}

func (a *aggregate) add(value float64) {
	a.count++
	if value < a.min {
		a.min = value
	}
	if value > a.max {
		a.max = value
	}
	a.avg = (a.avg*float64(a.count-1) + value) / float64(a.count)
	a.last = value
}

func (r *Repository) upsertSummary(ctx context.Context, tx *sql.Tx, deviceID string, agg *aggregate) error {
	var count int
	var minValue, maxValue, avgValue, lastValue float64
	err := tx.QueryRowContext(ctx, `SELECT sample_count, min_value, max_value, avg_value, last_value FROM telemetry_summaries WHERE device_id=? AND metric=? AND window_start=?`,
		deviceID, agg.metric, agg.windowStart).Scan(&count, &minValue, &maxValue, &avgValue, &lastValue)
	if err == nil {
		count += agg.count
		minValue = math.Min(minValue, agg.min)
		maxValue = math.Max(maxValue, agg.max)
		avgValue = (avgValue*float64(count-agg.count) + agg.avg*float64(agg.count)) / float64(count)
		lastValue = agg.last
		_, err = tx.ExecContext(ctx, `UPDATE telemetry_summaries SET sample_count=?, min_value=?, max_value=?, avg_value=?, last_value=?, updated_at=? WHERE device_id=? AND metric=? AND window_start=?`,
			count, minValue, maxValue, avgValue, lastValue, time.Now().UTC(), deviceID, agg.metric, agg.windowStart)
		return err
	}
	if err != sql.ErrNoRows {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO telemetry_summaries(device_id, metric, window_start, window_end, sample_count, min_value, max_value, avg_value, last_value, created_at, updated_at) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		deviceID, agg.metric, agg.windowStart, agg.windowStart.Add(time.Minute), agg.count, agg.min, agg.max, agg.avg, agg.last, time.Now().UTC(), time.Now().UTC())
	return err
}

func (r *Repository) Summaries(ctx context.Context, tenantID, deviceID string, since, until time.Time) ([]telemetrydomain.Summary, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT device_id, metric, window_start, window_end, sample_count, min_value, max_value, avg_value, last_value
		FROM telemetry_summaries WHERE device_id=? AND window_start>=? AND window_start<=? ORDER BY window_start ASC`, deviceID, since, until)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	summaries := make([]telemetrydomain.Summary, 0)
	for rows.Next() {
		var summary telemetrydomain.Summary
		if err := rows.Scan(&summary.DeviceID, &summary.Metric, &summary.WindowStart, &summary.WindowEnd, &summary.SampleCount, &summary.Min, &summary.Max, &summary.Avg, &summary.Last); err != nil {
			return nil, err
		}
		summaries = append(summaries, summary)
	}
	return summaries, rows.Err()
}

func (r *Repository) ListReadings(ctx context.Context, tenantID, deviceID string, query shareddomain.PageQuery) ([]telemetrydomain.ReadingRecord, int64, error) {
	where := "WHERE tenant_id=? AND device_id=?"
	args := []any{tenantID, deviceID}
	if v := query.Filters["metric"]; v != "" {
		where += " AND metric=?"
		args = append(args, v)
	}
	var total int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM telemetry_readings `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	sort := allowedTelemetrySort(query.Sort)
	listSQL := fmt.Sprintf(`SELECT device_id, metric, value, unit, received_at FROM telemetry_readings %s ORDER BY %s LIMIT ? OFFSET ?`, where, sort)
	args = append(args, query.Limit(), query.Offset())
	rows, err := r.db.QueryContext(ctx, listSQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	records := make([]telemetrydomain.ReadingRecord, 0)
	for rows.Next() {
		var record telemetrydomain.ReadingRecord
		if err := rows.Scan(&record.DeviceID, &record.Metric, &record.Value, &record.Unit, &record.ReceivedAt); err != nil {
			return nil, 0, err
		}
		records = append(records, record)
	}
	return records, total, rows.Err()
}

func (r *Repository) CountSince(ctx context.Context, tenantID, deviceID, metric string, since time.Time) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM telemetry_readings WHERE tenant_id=? AND device_id=? AND metric=? AND received_at>=?`, tenantID, deviceID, metric, since).Scan(&count)
	return count, err
}

func (r *Repository) DeleteSummariesBefore(ctx context.Context, before time.Time) (int64, error) {
	res, err := r.db.ExecContext(ctx, `DELETE FROM telemetry_summaries WHERE window_end<?`, before)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func allowedTelemetrySort(sort string) string {
	sort = strings.TrimSpace(sort)
	if sort == "" {
		return "received_at DESC"
	}
	field := strings.TrimPrefix(sort, "-")
	if field != "metric" && field != "received_at" {
		return "received_at DESC"
	}
	if strings.HasPrefix(sort, "-") {
		return field + " DESC"
	}
	return field + " ASC"
}
