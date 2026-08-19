package http

import "net/http"

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/tenants/{tenantID}/events", h.Create)
	mux.HandleFunc("GET /api/v1/tenants/{tenantID}/events", h.List)
	mux.HandleFunc("GET /api/v1/tenants/{tenantID}/events/{eventID}", h.Get)
	mux.HandleFunc("POST /api/v1/tenants/{tenantID}/events/{eventID}/acknowledge", h.Acknowledge)
}
