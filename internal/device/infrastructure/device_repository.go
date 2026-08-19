package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	devicedomain "github.com/acme/plantguard/internal/device/domain"
	shareddomain "github.com/acme/plantguard/internal/shared/domain"
	sharedinfra "github.com/acme/plantguard/internal/shared/infrastructure"
)

type DeviceRepository struct {
	db *sql.DB
}

func NewDeviceRepository(db *sql.DB) *DeviceRepository {
	return &DeviceRepository{db: db}
}

func (r *DeviceRepository) Create(ctx context.Context, device devicedomain.Device) error {
	tags, err := json.Marshal(device.Tags)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `INSERT INTO devices(id, tenant_id, site_id, model_id, name, protocol, enabled, firmware_version, tags, last_seen_at, status, created_at, updated_at, version)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		device.ID, device.TenantID, nullableString(device.SiteID), nullableString(device.ModelID), device.Name, device.Protocol, boolToInt(device.Enabled), device.FirmwareVersion, string(tags), nullableTime(device.LastSeenAt), device.Status, device.CreatedAt, device.UpdatedAt, device.Version)
	return sharedinfra.MapSQLError(err, nil)
}

func (r *DeviceRepository) GetByID(ctx context.Context, tenantID, id string) (devicedomain.Device, error) {
	device, err := r.get(ctx, `SELECT id, tenant_id, site_id, model_id, name, protocol, enabled, firmware_version, tags, last_seen_at, status, created_at, updated_at, version FROM devices WHERE tenant_id=? AND id=?`, tenantID, id)
	return device, sharedinfra.MapSQLError(err, nil)
}

func (r *DeviceRepository) Update(ctx context.Context, device devicedomain.Device) error {
	tags, err := json.Marshal(device.Tags)
	if err != nil {
		return err
	}
	res, err := r.db.ExecContext(ctx, `UPDATE devices SET site_id=?, model_id=?, name=?, protocol=?, enabled=?, firmware_version=?, tags=?, last_seen_at=?, status=?, updated_at=?, version=? WHERE tenant_id=? AND id=? AND version=?`,
		nullableString(device.SiteID), nullableString(device.ModelID), device.Name, device.Protocol, boolToInt(device.Enabled), device.FirmwareVersion, string(tags), nullableTime(device.LastSeenAt), device.Status, device.UpdatedAt, device.Version, device.TenantID, device.ID, device.Version-1)
	if err != nil {
		return sharedinfra.MapSQLError(err, nil)
	}
	return requireAffected(res)
}

func (r *DeviceRepository) UpdateHeartbeat(ctx context.Context, tenantID, deviceID string, observedAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `UPDATE devices SET last_seen_at=?, status='online', updated_at=? WHERE tenant_id=? AND id=?`, observedAt, time.Now().UTC(), tenantID, deviceID)
	return sharedinfra.MapSQLError(err, nil)
}

func (r *DeviceRepository) List(ctx context.Context, tenantID string, query shareddomain.PageQuery) ([]devicedomain.Device, int64, error) {
	where := "WHERE tenant_id=?"
	args := []any{tenantID}
	if v := query.Filters["name"]; v != "" {
		where += " AND name LIKE ?"
		args = append(args, "%"+v+"%")
	}
	if v := query.Filters["protocol"]; v != "" {
		where += " AND protocol=?"
		args = append(args, v)
	}
	if v := query.Filters["site_id"]; v != "" {
		where += " AND site_id=?"
		args = append(args, v)
	}
	var total int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM devices `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	sort := allowedDeviceSort(query.Sort)
	listSQL := fmt.Sprintf(`SELECT id, tenant_id, site_id, model_id, name, protocol, enabled, firmware_version, tags, last_seen_at, status, created_at, updated_at, version FROM devices %s ORDER BY %s LIMIT ? OFFSET ?`, where, sort)
	args = append(args, query.Limit(), query.Offset())
	rows, err := r.db.QueryContext(ctx, listSQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	devices := make([]devicedomain.Device, 0)
	for rows.Next() {
		device, err := scanDevice(rows)
		if err != nil {
			return nil, 0, err
		}
		devices = append(devices, device)
	}
	return devices, total, rows.Err()
}

func (r *DeviceRepository) ListByTenant(ctx context.Context, tenantID string) ([]devicedomain.Device, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, tenant_id, site_id, model_id, name, protocol, enabled, firmware_version, tags, last_seen_at, status, created_at, updated_at, version FROM devices WHERE tenant_id=?`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	devices := make([]devicedomain.Device, 0)
	for rows.Next() {
		device, err := scanDevice(rows)
		if err != nil {
			return nil, err
		}
		devices = append(devices, device)
	}
	return devices, rows.Err()
}

func (r *DeviceRepository) get(ctx context.Context, sql string, args ...any) (devicedomain.Device, error) {
	row := r.db.QueryRowContext(ctx, sql, args...)
	return scanDevice(row)
}

func scanDevice(row interface{ Scan(...any) error }) (devicedomain.Device, error) {
	var device devicedomain.Device
	var siteID, modelID, lastSeen sql.NullString
	var tagsRaw string
	var enabled int
	if err := row.Scan(&device.ID, &device.TenantID, &siteID, &modelID, &device.Name, &device.Protocol, &enabled, &device.FirmwareVersion, &tagsRaw, &lastSeen, &device.Status, &device.CreatedAt, &device.UpdatedAt, &device.Version); err != nil {
		return device, err
	}
	device.SiteID = siteID.String
	device.ModelID = modelID.String
	device.Enabled = enabled != 0
	if lastSeen.Valid {
		if t, err := time.Parse(time.RFC3339Nano, lastSeen.String); err == nil {
			device.LastSeenAt = t
		}
	}
	if err := json.Unmarshal([]byte(tagsRaw), &device.Tags); err != nil {
		device.Tags = []string{}
	}
	return device, nil
}

func allowedDeviceSort(sort string) string {
	sort = strings.TrimSpace(sort)
	if sort == "" {
		return "created_at DESC"
	}
	field := strings.TrimPrefix(sort, "-")
	if field != "name" && field != "protocol" && field != "created_at" && field != "status" {
		return "created_at DESC"
	}
	if strings.HasPrefix(sort, "-") {
		return field + " DESC"
	}
	return field + " ASC"
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func nullableTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func requireAffected(res sql.Result) error {
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return shareddomain.ErrPrecondition
	}
	return nil
}
