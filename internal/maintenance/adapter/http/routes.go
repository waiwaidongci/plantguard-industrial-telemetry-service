package http

import "net/http"

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/tenants/{tenantID}/maintenance/plans", h.CreatePlan)
	mux.HandleFunc("GET /api/v1/tenants/{tenantID}/maintenance/plans", h.ListPlans)
	mux.HandleFunc("GET /api/v1/tenants/{tenantID}/maintenance/plans/{planID}", h.GetPlan)
	mux.HandleFunc("PATCH /api/v1/tenants/{tenantID}/maintenance/plans/{planID}", h.UpdatePlan)
	mux.HandleFunc("GET /api/v1/tenants/{tenantID}/maintenance/tasks", h.ListTasks)
	mux.HandleFunc("GET /api/v1/tenants/{tenantID}/maintenance/tasks/{taskID}", h.GetTask)
	mux.HandleFunc("POST /api/v1/tenants/{tenantID}/maintenance/tasks/{taskID}/transition", h.TransitionTask)
}
