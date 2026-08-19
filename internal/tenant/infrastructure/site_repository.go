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

type SiteRepository struct {
	db *sql.DB
}

func NewSiteRepository(db *sql.DB) *SiteRepository {
	return &SiteRepository{db: db}
}

func (r *SiteRepository) Create(ctx context.Context, site domain.Site) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO sites(id, tenant_id, name, location, timezone, created_at, updated_at, version) VALUES(?, ?, ?, ?, ?, ?, ?, ?)`,
		site.ID, site.TenantID, site.Name, site.Location, site.Timezone, site.CreatedAt, site.UpdatedAt, site.Version)
	return sharedinfra.MapSQLError(err, nil)
}

func (r *SiteRepository) GetByID(ctx context.Context, tenantID, id string) (domain.Site, error) {
	var site domain.Site
	row := r.db.QueryRowContext(ctx, `SELECT id, tenant_id, name, location, timezone, created_at, updated_at, version FROM sites WHERE tenant_id=? AND id=?`, tenantID, id)
	err := row.Scan(&site.ID, &site.TenantID, &site.Name, &site.Location, &site.Timezone, &site.CreatedAt, &site.UpdatedAt, &site.Version)
	return site, sharedinfra.MapSQLError(err, nil)
}

func (r *SiteRepository) Update(ctx context.Context, site domain.Site) error {
	res, err := r.db.ExecContext(ctx, `UPDATE sites SET name=?, location=?, timezone=?, updated_at=?, version=? WHERE tenant_id=? AND id=? AND version=?`,
		site.Name, site.Location, site.Timezone, site.UpdatedAt, site.Version, site.TenantID, site.ID, site.Version-1)
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

func (r *SiteRepository) List(ctx context.Context, tenantID string, query shareddomain.PageQuery) ([]domain.Site, int64, error) {
	where := "WHERE tenant_id=?"
	args := []any{tenantID}
	if v := query.Filters["name"]; v != "" {
		where += " AND name LIKE ?"
		args = append(args, "%"+v+"%")
	}
	if v := query.Filters["location"]; v != "" {
		where += " AND location LIKE ?"
		args = append(args, "%"+v+"%")
	}
	var total int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM sites `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	sort := allowedSiteSort(query.Sort)
	listSQL := fmt.Sprintf(`SELECT id, tenant_id, name, location, timezone, created_at, updated_at, version FROM sites %s ORDER BY %s LIMIT ? OFFSET ?`, where, sort)
	args = append(args, query.Limit(), query.Offset())
	rows, err := r.db.QueryContext(ctx, listSQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	sites := make([]domain.Site, 0)
	for rows.Next() {
		var site domain.Site
		if err := rows.Scan(&site.ID, &site.TenantID, &site.Name, &site.Location, &site.Timezone, &site.CreatedAt, &site.UpdatedAt, &site.Version); err != nil {
			return nil, 0, err
		}
		sites = append(sites, site)
	}
	return sites, total, rows.Err()
}

func allowedSiteSort(sort string) string {
	sort = strings.TrimSpace(sort)
	if sort == "" {
		return "created_at DESC"
	}
	field := strings.TrimPrefix(sort, "-")
	switch field {
	case "name", "location", "created_at":
		if strings.HasPrefix(sort, "-") {
			return field + " DESC"
		}
		return field + " ASC"
	default:
		return "created_at DESC"
	}
}
