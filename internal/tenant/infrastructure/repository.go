package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	shareddomain "github.com/acme/plantguard/internal/shared/domain"
	sharedinfra "github.com/acme/plantguard/internal/shared/infrastructure"
	"github.com/acme/plantguard/internal/tenant/domain"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, tenant domain.Tenant) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO tenants(id, name, slug, created_at, updated_at, version) VALUES(?, ?, ?, ?, ?, ?)`,
		tenant.ID, tenant.Name, tenant.Slug, tenant.CreatedAt, tenant.UpdatedAt, tenant.Version)
	return sharedinfra.MapSQLError(err, nil)
}

func (r *Repository) GetByID(ctx context.Context, id string) (domain.Tenant, error) {
	return r.get(ctx, `SELECT id, name, slug, created_at, updated_at, version FROM tenants WHERE id=?`, id)
}

func (r *Repository) GetBySlug(ctx context.Context, slug string) (domain.Tenant, error) {
	return r.get(ctx, `SELECT id, name, slug, created_at, updated_at, version FROM tenants WHERE slug=?`, slug)
}

func (r *Repository) Update(ctx context.Context, tenant domain.Tenant) error {
	res, err := r.db.ExecContext(ctx, `UPDATE tenants SET name=?, slug=?, updated_at=?, version=? WHERE id=? AND version=?`,
		tenant.Name, tenant.Slug, tenant.UpdatedAt, tenant.Version, tenant.ID, tenant.Version-1)
	if err != nil {
		return sharedinfra.MapSQLError(err, nil)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return shareddomain.ErrPrecondition
	}
	return nil
}

func (r *Repository) List(ctx context.Context, query shareddomain.PageQuery) ([]domain.Tenant, int64, error) {
	where, args := tenantWhere(query.Filters)
	var total int64
	countSQL := `SELECT COUNT(*) FROM tenants ` + where
	if err := r.db.QueryRowContext(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	sort := allowedTenantSort(query.Sort)
	listSQL := fmt.Sprintf(`SELECT id, name, slug, created_at, updated_at, version FROM tenants %s ORDER BY %s LIMIT ? OFFSET ?`, where, sort)
	args = append(args, query.Limit(), query.Offset())
	rows, err := r.db.QueryContext(ctx, listSQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	tenants := make([]domain.Tenant, 0)
	for rows.Next() {
		tenant, err := scanTenant(rows)
		if err != nil {
			return nil, 0, err
		}
		tenants = append(tenants, tenant)
	}
	return tenants, total, rows.Err()
}

func (r *Repository) get(ctx context.Context, sql string, args ...any) (domain.Tenant, error) {
	var tenant domain.Tenant
	row := r.db.QueryRowContext(ctx, sql, args...)
	tenant, err := scanTenant(row)
	return tenant, sharedinfra.MapSQLError(err, nil)
}

type scanner interface {
	Scan(...any) error
}

func scanTenant(row scanner) (domain.Tenant, error) {
	var tenant domain.Tenant
	err := row.Scan(&tenant.ID, &tenant.Name, &tenant.Slug, &tenant.CreatedAt, &tenant.UpdatedAt, &tenant.Version)
	return tenant, err
}

func tenantWhere(filters map[string]string) (string, []any) {
	var clauses []string
	var args []any
	if v := filters["name"]; v != "" {
		clauses = append(clauses, "name LIKE ?")
		args = append(args, "%"+v+"%")
	}
	if v := filters["slug"]; v != "" {
		clauses = append(clauses, "slug LIKE ?")
		args = append(args, "%"+v+"%")
	}
	if len(clauses) == 0 {
		return "", args
	}
	return "WHERE " + strings.Join(clauses, " AND "), args
}

func allowedTenantSort(sort string) string {
	sort = strings.TrimSpace(sort)
	if sort == "" {
		return "created_at DESC"
	}
	field := strings.TrimPrefix(sort, "-")
	switch field {
	case "name", "slug", "created_at":
		if strings.HasPrefix(sort, "-") {
			return field + " DESC"
		}
		return field + " ASC"
	default:
		return "created_at DESC"
	}
}
