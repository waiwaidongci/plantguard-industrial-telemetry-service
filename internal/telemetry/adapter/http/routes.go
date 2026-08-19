package http

import "net/http"

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/tenants/{tenantID}/devices/{deviceID}/telemetry/batches", h.IngestBatch)
	mux.HandleFunc("GET /api/v1/tenants/{tenantID}/devices/{deviceID}/telemetry/summary", h.Summaries)
	mux.HandleFunc("GET /api/v1/tenants/{tenantID}/devices/{deviceID}/telemetry/readings", h.ListReadings)
}
