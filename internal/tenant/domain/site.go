package domain

import (
	"context"
	"time"

	shareddomain "github.com/acme/plantguard/internal/shared/domain"
)

type Site struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	Name      string    `json:"name"`
	Location  string    `json:"location"`
	Timezone  string    `json:"timezone"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Version   int64     `json:"version"`
}

func NewSite(tenantID, name, location, timezone string, now time.Time) Site {
	if timezone == "" {
		timezone = "UTC"
	}
	return Site{
		ID:        shareddomain.NewID("sit"),
		TenantID:  tenantID,
		Name:      name,
		Location:  location,
		Timezone:  timezone,
		CreatedAt: now,
		UpdatedAt: now,
		Version:   1,
	}
}

type SiteRepository interface {
	Create(context.Context, Site) error
	GetByID(context.Context, string, string) (Site, error)
	Update(context.Context, Site) error
	List(context.Context, string, shareddomain.PageQuery) ([]Site, int64, error)
}
