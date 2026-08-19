package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	ruledomain "github.com/acme/plantguard/internal/rule/domain"
	shareddomain "github.com/acme/plantguard/internal/shared/domain"
	sharedinfra "github.com/acme/plantguard/internal/shared/infrastructure"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, rule ruledomain.Rule) error {
	tags, err := rule.MarshalTagFilters()
	if err != nil {
		return err
	}
	actions, err := rule.MarshalActions()
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `INSERT INTO rules(id, tenant_id, name, description, metric, condition, threshold, duration_seconds, window_seconds, aggregation, tag_filters, severity, enabled, actions, created_at, updated_at, version)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		rule.ID, rule.TenantID, rule.Name, rule.Description, rule.Metric, rule.Condition, rule.Threshold, rule.DurationSeconds, rule.WindowSeconds, rule.Aggregation, tags, rule.Severity, boolToInt(rule.Enabled), actions, rule.CreatedAt, rule.UpdatedAt, rule.Version)
	return sharedinfra.MapSQLError(err, nil)
}

func (r *Repository) GetByID(ctx context.Context, tenantID, id string) (ruledomain.Rule, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, tenant_id, name, description, metric, condition, threshold, duration_seconds, window_seconds, aggregation, tag_filters, severity, enabled, actions, created_at, updated_at, version FROM rules WHERE tenant_id=? AND id=?`, tenantID, id)
	rule, err := scanRule(row)
	return rule, sharedinfra.MapSQLError(err, nil)
}

func (r *Repository) Update(ctx context.Context, rule ruledomain.Rule) error {
	tags, err := rule.MarshalTagFilters()
	if err != nil {
		return err
	}
	actions, err := rule.MarshalActions()
	if err != nil {
		return err
	}
	res, err := r.db.ExecContext(ctx, `UPDATE rules SET name=?, description=?, metric=?, condition=?, threshold=?, duration_seconds=?, window_seconds=?, aggregation=?, tag_filters=?, severity=?, enabled=?, actions=?, updated_at=?, version=? WHERE tenant_id=? AND id=? AND version=?`,
		rule.Name, rule.Description, rule.Metric, rule.Condition, rule.Threshold, rule.DurationSeconds, rule.WindowSeconds, rule.Aggregation, tags, rule.Severity, boolToInt(rule.Enabled), actions, rule.UpdatedAt, rule.Version, rule.TenantID, rule.ID, rule.Version-1)
	if err != nil {
		return sharedinfra.MapSQLError(err, nil)
	}
	return requireAffected(res)
}

func (r *Repository) List(ctx context.Context, tenantID string, query shareddomain.PageQuery) ([]ruledomain.Rule, int64, error) {
	where := "WHERE tenant_id=?"
	args := []any{tenantID}
	if v := query.Filters["metric"]; v != "" {
		where += " AND metric=?"
		args = append(args, v)
	}
	if v := query.Filters["name"]; v != "" {
		where += " AND name LIKE ?"
		args = append(args, "%"+v+"%")
	}
	var total int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM rules `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	sort := allowedRuleSort(query.Sort)
	listSQL := fmt.Sprintf(`SELECT id, tenant_id, name, description, metric, condition, threshold, duration_seconds, window_seconds, aggregation, tag_filters, severity, enabled, actions, created_at, updated_at, version FROM rules %s ORDER BY %s LIMIT ? OFFSET ?`, where, sort)
	args = append(args, query.Limit(), query.Offset())
	rows, err := r.db.QueryContext(ctx, listSQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	rules := make([]ruledomain.Rule, 0)
	for rows.Next() {
		rule, err := scanRule(rows)
		if err != nil {
			return nil, 0, err
		}
		rules = append(rules, rule)
	}
	return rules, total, rows.Err()
}

func (r *Repository) ListEnabledByTenant(ctx context.Context, tenantID string) ([]ruledomain.Rule, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, tenant_id, name, description, metric, condition, threshold, duration_seconds, window_seconds, aggregation, tag_filters, severity, enabled, actions, created_at, updated_at, version FROM rules WHERE tenant_id=? AND enabled=1`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	rules := make([]ruledomain.Rule, 0)
	for rows.Next() {
		rule, err := scanRule(rows)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return rules, rows.Err()
}

func scanRule(row interface{ Scan(...any) error }) (ruledomain.Rule, error) {
	var rule ruledomain.Rule
	var tags, actions string
	var enabled int
	if err := row.Scan(&rule.ID, &rule.TenantID, &rule.Name, &rule.Description, &rule.Metric, &rule.Condition, &rule.Threshold, &rule.DurationSeconds, &rule.WindowSeconds, &rule.Aggregation, &tags, &rule.Severity, &enabled, &actions, &rule.CreatedAt, &rule.UpdatedAt, &rule.Version); err != nil {
		return rule, err
	}
	rule.TagFilters = ruledomain.UnmarshalTagFilters(tags)
	rule.Actions = ruledomain.UnmarshalActions(actions)
	rule.Enabled = enabled != 0
	return rule, nil
}

func allowedRuleSort(sort string) string {
	sort = strings.TrimSpace(sort)
	if sort == "" {
		return "created_at DESC"
	}
	field := strings.TrimPrefix(sort, "-")
	if field != "name" && field != "metric" && field != "severity" && field != "created_at" {
		return "created_at DESC"
	}
	if strings.HasPrefix(sort, "-") {
		return field + " DESC"
	}
	return field + " ASC"
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
