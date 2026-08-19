package http

import "net/http"

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/tenants/{tenantID}/device-models", h.CreateModel)
	mux.HandleFunc("GET /api/v1/tenants/{tenantID}/device-models", h.ListModels)
	mux.HandleFunc("GET /api/v1/tenants/{tenantID}/device-models/{modelID}", h.GetModel)
	mux.HandleFunc("PATCH /api/v1/tenants/{tenantID}/device-models/{modelID}", h.UpdateModel)

	mux.HandleFunc("POST /api/v1/tenants/{tenantID}/devices", h.CreateDevice)
	mux.HandleFunc("GET /api/v1/tenants/{tenantID}/devices", h.ListDevices)
	mux.HandleFunc("GET /api/v1/tenants/{tenantID}/devices/{deviceID}", h.GetDevice)
	mux.HandleFunc("PATCH /api/v1/tenants/{tenantID}/devices/{deviceID}", h.UpdateDevice)
	mux.HandleFunc("POST /api/v1/tenants/{tenantID}/devices/{deviceID}/heartbeat", h.Heartbeat)
}
