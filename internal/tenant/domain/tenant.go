package domain

import (
	"context"
	"time"

	shareddomain "github.com/acme/plantguard/internal/shared/domain"
)

type Tenant struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Version   int64     `json:"version"`
}

func NewTenant(name, slug string, now time.Time) Tenant {
	if slug == "" {
		slug = slugFromName(name)
	}
	return Tenant{
		ID:        shareddomain.NewID("tnt"),
		Name:      name,
		Slug:      slug,
		CreatedAt: now,
		UpdatedAt: now,
		Version:   1,
	}
}

func (t Tenant) IsEmpty() bool {
	return t.ID == ""
}

type TenantRepository interface {
	Create(context.Context, Tenant) error
	GetByID(context.Context, string) (Tenant, error)
	GetBySlug(context.Context, string) (Tenant, error)
	Update(context.Context, Tenant) error
	List(context.Context, shareddomain.PageQuery) ([]Tenant, int64, error)
}
