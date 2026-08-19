package http

import "net/http"

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/tenants/{tenantID}/rules", h.CreateRule)
	mux.HandleFunc("GET /api/v1/tenants/{tenantID}/rules", h.ListRules)
	mux.HandleFunc("GET /api/v1/tenants/{tenantID}/rules/{ruleID}", h.GetRule)
	mux.HandleFunc("PATCH /api/v1/tenants/{tenantID}/rules/{ruleID}", h.UpdateRule)
}
