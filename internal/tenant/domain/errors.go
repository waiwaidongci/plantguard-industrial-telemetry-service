package domain

import shareddomain "github.com/acme/plantguard/internal/shared/domain"

var (
	ErrTenantNameRequired = shareddomain.New("tenant_name_required", "tenant name is required", 400)
	ErrTenantSlugRequired = shareddomain.New("tenant_slug_required", "tenant slug is required", 400)
	ErrSiteNameRequired   = shareddomain.New("site_name_required", "site name is required", 400)
	ErrInvalidVersion     = shareddomain.ErrPrecondition
)
