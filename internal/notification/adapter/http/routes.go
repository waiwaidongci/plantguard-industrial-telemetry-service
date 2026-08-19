package http

import "net/http"

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/tenants/{tenantID}/notifications", h.Send)
	mux.HandleFunc("GET /api/v1/tenants/{tenantID}/notifications", h.List)
}
