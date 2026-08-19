package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	maintenancedomain "github.com/acme/plantguard/internal/maintenance/domain"
	shareddomain "github.com/acme/plantguard/internal/shared/domain"
	sharedinfra "github.com/acme/plantguard/internal/shared/infrastructure"
)

type PlanRepository struct {
	db *sql.DB
}

type TaskRepository struct {
	db *sql.DB
}

func NewPlanRepository(db *sql.DB) *PlanRepository {
	return &PlanRepository{db: db}
}

func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *PlanRepository) Create(ctx context.Context, plan maintenancedomain.Plan) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO maintenance_plans(id, tenant_id, device_id, name, trigger_type, interval_hours, calendar_spec, event_type, status, next_due_at, created_at, updated_at, version)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		plan.ID, plan.TenantID, nullableString(plan.DeviceID), plan.Name, plan.TriggerType, plan.IntervalHours, plan.CalendarSpec, plan.EventType, plan.Status, nullableTime(plan.NextDueAt), plan.CreatedAt, plan.UpdatedAt, plan.Version)
	return sharedinfra.MapSQLError(err, nil)
}

func (r *PlanRepository) GetByID(ctx context.Context, tenantID, id string) (maintenancedomain.Plan, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, tenant_id, device_id, name, trigger_type, interval_hours, calendar_spec, event_type, status, next_due_at, created_at, updated_at, version FROM maintenance_plans WHERE tenant_id=? AND id=?`, tenantID, id)
	plan, err := scanPlan(row)
	return plan, sharedinfra.MapSQLError(err, nil)
}

func (r *PlanRepository) Update(ctx context.Context, plan maintenancedomain.Plan) error {
	res, err := r.db.ExecContext(ctx, `UPDATE maintenance_plans SET device_id=?, name=?, trigger_type=?, interval_hours=?, calendar_spec=?, event_type=?, status=?, next_due_at=?, updated_at=?, version=? WHERE tenant_id=? AND id=? AND version=?`,
		nullableString(plan.DeviceID), plan.Name, plan.TriggerType, plan.IntervalHours, plan.CalendarSpec, plan.EventType, plan.Status, nullableTime(plan.NextDueAt), plan.UpdatedAt, plan.Version, plan.TenantID, plan.ID, plan.Version-1)
	if err != nil {
		return sharedinfra.MapSQLError(err, nil)
	}
	return requireAffected(res)
}

func (r *PlanRepository) List(ctx context.Context, tenantID string, query shareddomain.PageQuery) ([]maintenancedomain.Plan, int64, error) {
	where := "WHERE tenant_id=?"
	args := []any{tenantID}
	if v := query.Filters["status"]; v != "" {
		where += " AND status=?"
		args = append(args, v)
	}
	if v := query.Filters["device_id"]; v != "" {
		where += " AND device_id=?"
		args = append(args, v)
	}
	var total int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM maintenance_plans `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	sort := allowedPlanSort(query.Sort)
	listSQL := fmt.Sprintf(`SELECT id, tenant_id, device_id, name, trigger_type, interval_hours, calendar_spec, event_type, status, next_due_at, created_at, updated_at, version FROM maintenance_plans %s ORDER BY %s LIMIT ? OFFSET ?`, where, sort)
	args = append(args, query.Limit(), query.Offset())
	rows, err := r.db.QueryContext(ctx, listSQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	plans := make([]maintenancedomain.Plan, 0)
	for rows.Next() {
		plan, err := scanPlan(rows)
		if err != nil {
			return nil, 0, err
		}
		plans = append(plans, plan)
	}
	return plans, total, rows.Err()
}

func (r *PlanRepository) ListActiveDue(ctx context.Context, now time.Time) ([]maintenancedomain.Plan, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, tenant_id, device_id, name, trigger_type, interval_hours, calendar_spec, event_type, status, next_due_at, created_at, updated_at, version FROM maintenance_plans WHERE status='active' AND next_due_at IS NOT NULL AND next_due_at<=? ORDER BY next_due_at ASC`, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	plans := make([]maintenancedomain.Plan, 0)
	for rows.Next() {
		plan, err := scanPlan(rows)
		if err != nil {
			return nil, err
		}
		plans = append(plans, plan)
	}
	return plans, rows.Err()
}

func (r *TaskRepository) Create(ctx context.Context, task maintenancedomain.Task) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO maintenance_tasks(id, tenant_id, plan_id, device_id, status, due_at, completed_at, notes, created_at, updated_at, version)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		task.ID, task.TenantID, nullableString(task.PlanID), nullableString(task.DeviceID), task.Status, nullableTime(task.DueAt), nullableTime(task.CompletedAt), task.Notes, task.CreatedAt, task.UpdatedAt, task.Version)
	return sharedinfra.MapSQLError(err, nil)
}

func (r *TaskRepository) GetByID(ctx context.Context, tenantID, id string) (maintenancedomain.Task, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, tenant_id, plan_id, device_id, status, due_at, completed_at, notes, created_at, updated_at, version FROM maintenance_tasks WHERE tenant_id=? AND id=?`, tenantID, id)
	task, err := scanTask(row)
	return task, sharedinfra.MapSQLError(err, nil)
}

func (r *TaskRepository) Update(ctx context.Context, task maintenancedomain.Task) error {
	res, err := r.db.ExecContext(ctx, `UPDATE maintenance_tasks SET plan_id=?, device_id=?, status=?, due_at=?, completed_at=?, notes=?, updated_at=?, version=? WHERE tenant_id=? AND id=? AND version=?`,
		nullableString(task.PlanID), nullableString(task.DeviceID), task.Status, nullableTime(task.DueAt), nullableTime(task.CompletedAt), task.Notes, task.UpdatedAt, task.Version, task.TenantID, task.ID, task.Version-1)
	if err != nil {
		return sharedinfra.MapSQLError(err, nil)
	}
	return requireAffected(res)
}

func (r *TaskRepository) List(ctx context.Context, tenantID string, query shareddomain.PageQuery) ([]maintenancedomain.Task, int64, error) {
	where := "WHERE tenant_id=?"
	args := []any{tenantID}
	if v := query.Filters["status"]; v != "" {
		where += " AND status=?"
		args = append(args, v)
	}
	if v := query.Filters["device_id"]; v != "" {
		where += " AND device_id=?"
		args = append(args, v)
	}
	var total int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM maintenance_tasks `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	sort := allowedTaskSort(query.Sort)
	listSQL := fmt.Sprintf(`SELECT id, tenant_id, plan_id, device_id, status, due_at, completed_at, notes, created_at, updated_at, version FROM maintenance_tasks %s ORDER BY %s LIMIT ? OFFSET ?`, where, sort)
	args = append(args, query.Limit(), query.Offset())
	rows, err := r.db.QueryContext(ctx, listSQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	tasks := make([]maintenancedomain.Task, 0)
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, 0, err
		}
		tasks = append(tasks, task)
	}
	return tasks, total, rows.Err()
}

func (r *TaskRepository) ListOpenByDevice(ctx context.Context, tenantID, deviceID string) ([]maintenancedomain.Task, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, tenant_id, plan_id, device_id, status, due_at, completed_at, notes, created_at, updated_at, version FROM maintenance_tasks WHERE tenant_id=? AND device_id=? AND status IN ('open','in_progress','retrying') ORDER BY due_at ASC`, tenantID, deviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tasks := make([]maintenancedomain.Task, 0)
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}

func scanPlan(row interface{ Scan(...any) error }) (maintenancedomain.Plan, error) {
	var plan maintenancedomain.Plan
	var deviceID, nextDue sql.NullString
	if err := row.Scan(&plan.ID, &plan.TenantID, &deviceID, &plan.Name, &plan.TriggerType, &plan.IntervalHours, &plan.CalendarSpec, &plan.EventType, &plan.Status, &nextDue, &plan.CreatedAt, &plan.UpdatedAt, &plan.Version); err != nil {
		return plan, err
	}
	plan.DeviceID = deviceID.String
	if nextDue.Valid {
		if t, err := time.Parse(time.RFC3339Nano, nextDue.String); err == nil {
			plan.NextDueAt = t
		}
	}
	return plan, nil
}

func scanTask(row interface{ Scan(...any) error }) (maintenancedomain.Task, error) {
	var task maintenancedomain.Task
	var planID, deviceID, dueAt, completedAt sql.NullString
	if err := row.Scan(&task.ID, &task.TenantID, &planID, &deviceID, &task.Status, &dueAt, &completedAt, &task.Notes, &task.CreatedAt, &task.UpdatedAt, &task.Version); err != nil {
		return task, err
	}
	task.PlanID = planID.String
	task.DeviceID = deviceID.String
	if dueAt.Valid {
		if t, err := time.Parse(time.RFC3339Nano, dueAt.String); err == nil {
			task.DueAt = t
		}
	}
	if completedAt.Valid {
		if t, err := time.Parse(time.RFC3339Nano, completedAt.String); err == nil {
			task.CompletedAt = t
		}
	}
	return task, nil
}

func allowedPlanSort(sort string) string {
	sort = strings.TrimSpace(sort)
	if sort == "" {
		return "created_at DESC"
	}
	field := strings.TrimPrefix(sort, "-")
	if field != "name" && field != "status" && field != "next_due_at" && field != "created_at" {
		return "created_at DESC"
	}
	if strings.HasPrefix(sort, "-") {
		return field + " DESC"
	}
	return field + " ASC"
}

func allowedTaskSort(sort string) string {
	sort = strings.TrimSpace(sort)
	if sort == "" {
		return "created_at DESC"
	}
	field := strings.TrimPrefix(sort, "-")
	if field != "status" && field != "due_at" && field != "created_at" {
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
