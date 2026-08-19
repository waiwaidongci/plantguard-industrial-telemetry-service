package http

import (
	"net/http"
	"strconv"

	sharedhttp "github.com/acme/plantguard/internal/shared/adapter/http"
	shareddomain "github.com/acme/plantguard/internal/shared/domain"
	"github.com/acme/plantguard/internal/tenant/application"
	tenantdomain "github.com/acme/plantguard/internal/tenant/domain"
)

type Handler struct {
	service application.Service
}

func NewHandler(service application.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateTenant(w http.ResponseWriter, r *http.Request) {
	var input application.CreateTenantInput
	if err := sharedhttp.DecodeJSON(r, &input); err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	tenant, created, err := h.service.CreateTenant(r.Context(), r.Header.Get("Idempotency-Key"), input)
	if err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	status := http.StatusCreated
	if !created {
		status = http.StatusOK
	}
	sharedhttp.WriteJSON(w, status, sharedhttp.Envelope{Data: tenant})
}

func (h *Handler) GetTenant(w http.ResponseWriter, r *http.Request) {
	tenant, err := h.service.GetTenant(r.Context(), sharedhttp.PathValue(r, "tenantID"))
	if err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	sharedhttp.WriteOK(w, tenant)
}

func (h *Handler) UpdateTenant(w http.ResponseWriter, r *http.Request) {
	var input application.UpdateTenantInput
	if err := sharedhttp.DecodeJSON(r, &input); err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	if raw := r.Header.Get("If-Match"); raw != "" {
		v, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			sharedhttp.WriteError(w, r, shareddomain.ErrBadRequest)
			return
		}
		input.Version = v
	}
	tenant, err := h.service.UpdateTenant(r.Context(), sharedhttp.PathValue(r, "tenantID"), input)
	if err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	sharedhttp.WriteOK(w, tenant)
}

func (h *Handler) ListTenants(w http.ResponseWriter, r *http.Request) {
	query, err := shareddomain.ParsePageQuery(r.URL.Query())
	if err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	items, total, err := h.service.ListTenants(r.Context(), query)
	if err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	sharedhttp.WriteJSON(w, http.StatusOK, sharedhttp.Envelope{Data: items, Meta: shareddomain.NewPage(toAny(items), query, total)})
}

func (h *Handler) CreateSite(w http.ResponseWriter, r *http.Request) {
	var input application.CreateSiteInput
	if err := sharedhttp.DecodeJSON(r, &input); err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	site, created, err := h.service.CreateSite(r.Context(), sharedhttp.PathValue(r, "tenantID"), r.Header.Get("Idempotency-Key"), input)
	if err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	status := http.StatusCreated
	if !created {
		status = http.StatusOK
	}
	sharedhttp.WriteJSON(w, status, sharedhttp.Envelope{Data: site})
}

func (h *Handler) GetSite(w http.ResponseWriter, r *http.Request) {
	site, err := h.service.GetSite(r.Context(), sharedhttp.PathValue(r, "tenantID"), sharedhttp.PathValue(r, "siteID"))
	if err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	sharedhttp.WriteOK(w, site)
}

func (h *Handler) UpdateSite(w http.ResponseWriter, r *http.Request) {
	var input application.UpdateSiteInput
	if err := sharedhttp.DecodeJSON(r, &input); err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	if raw := r.Header.Get("If-Match"); raw != "" {
		v, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			sharedhttp.WriteError(w, r, shareddomain.ErrBadRequest)
			return
		}
		input.Version = v
	}
	site, err := h.service.UpdateSite(r.Context(), sharedhttp.PathValue(r, "tenantID"), sharedhttp.PathValue(r, "siteID"), input)
	if err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	sharedhttp.WriteOK(w, site)
}

func (h *Handler) ListSites(w http.ResponseWriter, r *http.Request) {
	query, err := shareddomain.ParsePageQuery(r.URL.Query())
	if err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	items, total, err := h.service.ListSites(r.Context(), sharedhttp.PathValue(r, "tenantID"), query)
	if err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	sharedhttp.WriteJSON(w, http.StatusOK, sharedhttp.Envelope{Data: items, Meta: shareddomain.NewPage(toAnySite(items), query, total)})
}

func toAny(items []tenantdomain.Tenant) []any {
	out := make([]any, len(items))
	for i := range items {
		out[i] = items[i]
	}
	return out
}

func toAnySite(items []tenantdomain.Site) []any {
	out := make([]any, len(items))
	for i := range items {
		out[i] = items[i]
	}
	return out
}
