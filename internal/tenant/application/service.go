package application

import (
	"context"
	"strings"

	sharedapplication "github.com/acme/plantguard/internal/shared/application"
	shareddomain "github.com/acme/plantguard/internal/shared/domain"
	sharedinfra "github.com/acme/plantguard/internal/shared/infrastructure"
	"github.com/acme/plantguard/internal/tenant/domain"
)

type Service interface {
	CreateTenant(ctx context.Context, idempotencyKey string, input CreateTenantInput) (domain.Tenant, bool, error)
	GetTenant(ctx context.Context, id string) (domain.Tenant, error)
	UpdateTenant(ctx context.Context, id string, input UpdateTenantInput) (domain.Tenant, error)
	ListTenants(ctx context.Context, query shareddomain.PageQuery) ([]domain.Tenant, int64, error)
	CreateSite(ctx context.Context, tenantID, idempotencyKey string, input CreateSiteInput) (domain.Site, bool, error)
	GetSite(ctx context.Context, tenantID, id string) (domain.Site, error)
	UpdateSite(ctx context.Context, tenantID, id string, input UpdateSiteInput) (domain.Site, error)
	ListSites(ctx context.Context, tenantID string, query shareddomain.PageQuery) ([]domain.Site, int64, error)
}

type service struct {
	tenants     domain.TenantRepository
	sites       domain.SiteRepository
	idempotency sharedapplication.IdempotencyStore
	clock       sharedinfra.Clock
}

func NewService(tenants domain.TenantRepository, sites domain.SiteRepository, idempotency sharedapplication.IdempotencyStore, clock sharedinfra.Clock) Service {
	return &service{tenants: tenants, sites: sites, idempotency: idempotency, clock: clock}
}

func (s *service) CreateTenant(ctx context.Context, idempotencyKey string, input CreateTenantInput) (domain.Tenant, bool, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return domain.Tenant{}, false, domain.ErrTenantNameRequired
	}
	if idempotencyKey != "" {
		record, found, err := s.idempotency.Get(ctx, "", idempotencyKey)
		if err != nil {
			return domain.Tenant{}, false, err
		}
		if found {
			tenant, err := s.tenants.GetByID(ctx, record.ResourceID)
			return tenant, false, err
		}
	}
	now := s.clock.Now(ctx)
	tenant := domain.NewTenant(name, input.Slug, now)
	if err := s.tenants.Create(ctx, tenant); err != nil {
		return domain.Tenant{}, false, err
	}
	if idempotencyKey != "" {
		if err := s.idempotency.Put(ctx, sharedapplication.IdempotencyRecord{
			TenantID: "", Key: idempotencyKey, ResourceKind: "tenant", ResourceID: tenant.ID,
		}); err != nil {
			return domain.Tenant{}, true, err
		}
	}
	return tenant, true, nil
}

func (s *service) GetTenant(ctx context.Context, id string) (domain.Tenant, error) {
	return s.tenants.GetByID(ctx, id)
}

func (s *service) UpdateTenant(ctx context.Context, id string, input UpdateTenantInput) (domain.Tenant, error) {
	tenant, err := s.tenants.GetByID(ctx, id)
	if err != nil {
		return domain.Tenant{}, err
	}
	if input.Version != 0 && input.Version != tenant.Version {
		return domain.Tenant{}, domain.ErrInvalidVersion
	}
	if strings.TrimSpace(input.Name) != "" {
		tenant.Name = strings.TrimSpace(input.Name)
	}
	if strings.TrimSpace(input.Slug) != "" {
		tenant.Slug = strings.TrimSpace(input.Slug)
	}
	tenant.UpdatedAt = s.clock.Now(ctx)
	tenant.Version++
	if err := s.tenants.Update(ctx, tenant); err != nil {
		return domain.Tenant{}, err
	}
	return tenant, nil
}

func (s *service) ListTenants(ctx context.Context, query shareddomain.PageQuery) ([]domain.Tenant, int64, error) {
	return s.tenants.List(ctx, query)
}

func (s *service) CreateSite(ctx context.Context, tenantID, idempotencyKey string, input CreateSiteInput) (domain.Site, bool, error) {
	if err := validateSite(CreateSiteInput{Name: input.Name, Location: input.Location, Timezone: "UTC"}); err != nil {
		return domain.Site{}, false, err
	}
	if _, err := s.tenants.GetByID(ctx, tenantID); err != nil {
		return domain.Site{}, false, err
	}
	if idempotencyKey != "" {
		record, found, err := s.idempotency.Get(ctx, tenantID, idempotencyKey)
		if err != nil {
			return domain.Site{}, false, err
		}
		if found {
			site, err := s.sites.GetByID(ctx, tenantID, record.ResourceID)
			return site, false, err
		}
	}
	site := domain.NewSite(tenantID, strings.TrimSpace(input.Name), strings.TrimSpace(input.Location), strings.TrimSpace(input.Timezone), s.clock.Now(ctx))
	if err := s.sites.Create(ctx, site); err != nil {
		return domain.Site{}, false, err
	}
	if idempotencyKey != "" {
		if err := s.idempotency.Put(ctx, sharedapplication.IdempotencyRecord{
			TenantID: tenantID, Key: idempotencyKey, ResourceKind: "site", ResourceID: site.ID,
		}); err != nil {
			return domain.Site{}, true, err
		}
	}
	return site, true, nil
}

func (s *service) GetSite(ctx context.Context, tenantID, id string) (domain.Site, error) {
	return s.sites.GetByID(ctx, tenantID, id)
}

func (s *service) UpdateSite(ctx context.Context, tenantID, id string, input UpdateSiteInput) (domain.Site, error) {
	site, err := s.sites.GetByID(ctx, tenantID, id)
	if err != nil {
		return domain.Site{}, err
	}
	if input.Version != 0 && input.Version != site.Version {
		return domain.Site{}, domain.ErrInvalidVersion
	}
	if strings.TrimSpace(input.Name) != "" {
		site.Name = strings.TrimSpace(input.Name)
	}
	if strings.TrimSpace(input.Location) != "" {
		site.Location = strings.TrimSpace(input.Location)
	}
	if strings.TrimSpace(input.Timezone) != "" {
		site.Timezone = strings.TrimSpace(input.Timezone)
	}
	site.UpdatedAt = s.clock.Now(ctx)
	site.Version++
	if err := s.sites.Update(ctx, site); err != nil {
		return domain.Site{}, err
	}
	return site, nil
}

func (s *service) ListSites(ctx context.Context, tenantID string, query shareddomain.PageQuery) ([]domain.Site, int64, error) {
	return s.sites.List(ctx, tenantID, query)
}
