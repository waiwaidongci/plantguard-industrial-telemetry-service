package http

import "net/http"

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/tenants", h.CreateTenant)
	mux.HandleFunc("GET /api/v1/tenants", h.ListTenants)
	mux.HandleFunc("GET /api/v1/tenants/{tenantID}", h.GetTenant)
	mux.HandleFunc("PATCH /api/v1/tenants/{tenantID}", h.UpdateTenant)
	mux.HandleFunc("POST /api/v1/tenants/{tenantID}/sites", h.CreateSite)
	mux.HandleFunc("GET /api/v1/tenants/{tenantID}/sites", h.ListSites)
	mux.HandleFunc("GET /api/v1/tenants/{tenantID}/sites/{siteID}", h.GetSite)
	mux.HandleFunc("PATCH /api/v1/tenants/{tenantID}/sites/{siteID}", h.UpdateSite)
}
