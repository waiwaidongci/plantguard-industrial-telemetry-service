package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	eventdomain "github.com/acme/plantguard/internal/event/domain"
	shareddomain "github.com/acme/plantguard/internal/shared/domain"
	sharedinfra "github.com/acme/plantguard/internal/shared/infrastructure"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, event eventdomain.Event) error {
	data, err := event.MarshalData()
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `INSERT INTO events(id, tenant_id, device_id, rule_id, type, severity, message, data, status, occurred_at, acknowledged_by, acknowledged_at, created_at, updated_at, version)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		event.ID, event.TenantID, event.DeviceID, nullableString(event.RuleID), event.Type, event.Severity, event.Message, data, event.Status, event.OccurredAt, event.AcknowledgedBy, nullableTime(event.AcknowledgedAt), event.CreatedAt, event.UpdatedAt, event.Version)
	return sharedinfra.MapSQLError(err, nil)
}

func (r *Repository) GetByID(ctx context.Context, tenantID, id string) (eventdomain.Event, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, tenant_id, device_id, rule_id, type, severity, message, data, status, occurred_at, acknowledged_by, acknowledged_at, created_at, updated_at, version FROM events WHERE tenant_id=? AND id=?`, tenantID, id)
	event, err := scanEvent(row)
	return event, sharedinfra.MapSQLError(err, nil)
}

func (r *Repository) Update(ctx context.Context, event eventdomain.Event) error {
	data, err := event.MarshalData()
	if err != nil {
		return err
	}
	res, err := r.db.ExecContext(ctx, `UPDATE events SET rule_id=?, type=?, severity=?, message=?, data=?, status=?, occurred_at=?, acknowledged_by=?, acknowledged_at=?, updated_at=?, version=? WHERE tenant_id=? AND id=? AND version=?`,
		nullableString(event.RuleID), event.Type, event.Severity, event.Message, data, event.Status, event.OccurredAt, event.AcknowledgedBy, nullableTime(event.AcknowledgedAt), event.UpdatedAt, event.Version, event.TenantID, event.ID, event.Version-1)
	if err != nil {
		return sharedinfra.MapSQLError(err, nil)
	}
	return requireAffected(res)
}

func (r *Repository) List(ctx context.Context, tenantID string, query shareddomain.PageQuery) ([]eventdomain.Event, int64, error) {
	where := "WHERE tenant_id=?"
	args := []any{tenantID}
	if v := query.Filters["device_id"]; v != "" {
		where += " AND device_id=?"
		args = append(args, v)
	}
	if v := query.Filters["type"]; v != "" {
		where += " AND type=?"
		args = append(args, v)
	}
	if v := query.Filters["severity"]; v != "" {
		where += " AND severity=?"
		args = append(args, v)
	}
	if v := query.Filters["status"]; v != "" {
		where += " AND status=?"
		args = append(args, v)
	}
	var total int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM events `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	sort := allowedEventSort(query.Sort)
	listSQL := fmt.Sprintf(`SELECT id, tenant_id, device_id, rule_id, type, severity, message, data, status, occurred_at, acknowledged_by, acknowledged_at, created_at, updated_at, version FROM events %s ORDER BY %s LIMIT ? OFFSET ?`, where, sort)
	args = append(args, query.Limit(), query.Offset())
	rows, err := r.db.QueryContext(ctx, listSQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	events := make([]eventdomain.Event, 0)
	for rows.Next() {
		event, err := scanEvent(rows)
		if err != nil {
			return nil, 0, err
		}
		events = append(events, event)
	}
	return events, total, rows.Err()
}

func (r *Repository) ListOpenByDeviceSince(ctx context.Context, tenantID, deviceID, eventType string, since time.Time) ([]eventdomain.Event, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, tenant_id, device_id, rule_id, type, severity, message, data, status, occurred_at, acknowledged_by, acknowledged_at, created_at, updated_at, version FROM events WHERE tenant_id=? AND device_id=? AND status='open' AND type=? AND occurred_at>=? ORDER BY occurred_at DESC`, tenantID, deviceID, eventType, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	events := make([]eventdomain.Event, 0)
	for rows.Next() {
		event, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

func scanEvent(row interface{ Scan(...any) error }) (eventdomain.Event, error) {
	var event eventdomain.Event
	var ruleID sql.NullString
	var acknowledgedAt sql.NullString
	var data string
	if err := row.Scan(&event.ID, &event.TenantID, &event.DeviceID, &ruleID, &event.Type, &event.Severity, &event.Message, &data, &event.Status, &event.OccurredAt, &event.AcknowledgedBy, &acknowledgedAt, &event.CreatedAt, &event.UpdatedAt, &event.Version); err != nil {
		return event, err
	}
	event.RuleID = ruleID.String
	event.Data = eventdomain.UnmarshalData(data)
	if acknowledgedAt.Valid {
		if t, err := time.Parse(time.RFC3339Nano, acknowledgedAt.String); err == nil {
			event.AcknowledgedAt = t
		}
	}
	return event, nil
}

func allowedEventSort(sort string) string {
	sort = strings.TrimSpace(sort)
	if sort == "" {
		return "occurred_at DESC"
	}
	field := strings.TrimPrefix(sort, "-")
	if field != "occurred_at" && field != "severity" && field != "status" && field != "created_at" {
		return "occurred_at DESC"
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
