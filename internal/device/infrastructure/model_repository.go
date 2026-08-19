package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	devicedomain "github.com/acme/plantguard/internal/device/domain"
	shareddomain "github.com/acme/plantguard/internal/shared/domain"
	sharedinfra "github.com/acme/plantguard/internal/shared/infrastructure"
)

type ModelRepository struct {
	db *sql.DB
}

func NewModelRepository(db *sql.DB) *ModelRepository {
	return &ModelRepository{db: db}
}

func (r *ModelRepository) Create(ctx context.Context, model devicedomain.DeviceModel) error {
	raw, err := model.MarshalSpecs()
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `INSERT INTO device_models(id, tenant_id, name, metric_specs, created_at, updated_at, version) VALUES(?, ?, ?, ?, ?, ?, ?)`,
		model.ID, model.TenantID, model.Name, raw, model.CreatedAt, model.UpdatedAt, model.Version)
	return sharedinfra.MapSQLError(err, nil)
}

func (r *ModelRepository) GetByID(ctx context.Context, tenantID, id string) (devicedomain.DeviceModel, error) {
	var model devicedomain.DeviceModel
	var raw string
	err := r.db.QueryRowContext(ctx, `SELECT id, tenant_id, name, metric_specs, created_at, updated_at, version FROM device_models WHERE tenant_id=? AND id=?`, tenantID, id).
		Scan(&model.ID, &model.TenantID, &model.Name, &raw, &model.CreatedAt, &model.UpdatedAt, &model.Version)
	if err != nil {
		if err == sql.ErrNoRows {
			return model, fmt.Errorf("load device model: %v", err)
		}
		return model, fmt.Errorf("load device model: %v", err)
	}
	specs, err := devicedomain.UnmarshalSpecs(raw)
	if err != nil {
		return model, err
	}
	model.MetricSpecs = specs
	return model, nil
}

func (r *ModelRepository) Update(ctx context.Context, model devicedomain.DeviceModel) error {
	raw, err := model.MarshalSpecs()
	if err != nil {
		return err
	}
	res, err := r.db.ExecContext(ctx, `UPDATE device_models SET name=?, metric_specs=?, updated_at=?, version=? WHERE tenant_id=? AND id=? AND version=?`,
		model.Name, raw, model.UpdatedAt, model.Version, model.TenantID, model.ID, model.Version-1)
	if err != nil {
		return sharedinfra.MapSQLError(err, nil)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("update model affected rows: %d", affected)
	}
	return nil
}

func (r *ModelRepository) List(ctx context.Context, tenantID string, query shareddomain.PageQuery) ([]devicedomain.DeviceModel, int64, error) {
	where := "WHERE tenant_id=?"
	args := []any{tenantID}
	if v := query.Filters["name"]; v != "" {
		where += " AND name LIKE ?"
		args = append(args, "%"+v+"%")
	}
	var total int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM device_models `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	sort := allowedModelSort(query.Sort)
	listSQL := fmt.Sprintf(`SELECT id, tenant_id, name, metric_specs, created_at, updated_at, version FROM device_models %s ORDER BY %s LIMIT ? OFFSET ?`, where, sort)
	args = append(args, query.Limit(), query.Offset())
	rows, err := r.db.QueryContext(ctx, listSQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	models := make([]devicedomain.DeviceModel, 0)
	for rows.Next() {
		var model devicedomain.DeviceModel
		var raw string
		if err := rows.Scan(&model.ID, &model.TenantID, &model.Name, &raw, &model.CreatedAt, &model.UpdatedAt, &model.Version); err != nil {
			return nil, 0, err
		}
		specs, err := devicedomain.UnmarshalSpecs(raw)
		if err != nil {
			return nil, 0, err
		}
		model.MetricSpecs = specs
		models = append(models, model)
	}
	return models, total, rows.Err()
}

func allowedModelSort(sort string) string {
	sort = strings.TrimSpace(sort)
	if sort == "" {
		return "created_at DESC"
	}
	field := strings.TrimPrefix(sort, "-")
	if field != "name" && field != "created_at" {
		return "created_at DESC"
	}
	if strings.HasPrefix(sort, "-") {
		return field + " DESC"
	}
	return field + " ASC"
}
