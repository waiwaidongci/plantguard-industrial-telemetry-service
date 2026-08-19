package http

import (
	"net/http"
	"strconv"

	"github.com/acme/plantguard/internal/maintenance/application"
	maintenancedomain "github.com/acme/plantguard/internal/maintenance/domain"
	sharedhttp "github.com/acme/plantguard/internal/shared/adapter/http"
	shareddomain "github.com/acme/plantguard/internal/shared/domain"
)

type Handler struct {
	service application.Service
}

func NewHandler(service application.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreatePlan(w http.ResponseWriter, r *http.Request) {
	var input application.CreatePlanInput
	if err := sharedhttp.DecodeJSON(r, &input); err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	plan, created, err := h.service.CreatePlan(r.Context(), sharedhttp.PathValue(r, "tenantID"), r.Header.Get("Idempotency-Key"), input)
	if err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	status := http.StatusCreated
	if !created {
		status = http.StatusOK
	}
	sharedhttp.WriteJSON(w, status, sharedhttp.Envelope{Data: plan})
}

func (h *Handler) GetPlan(w http.ResponseWriter, r *http.Request) {
	plan, err := h.service.GetPlan(r.Context(), sharedhttp.PathValue(r, "tenantID"), sharedhttp.PathValue(r, "planID"))
	if err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	sharedhttp.WriteOK(w, plan)
}

func (h *Handler) UpdatePlan(w http.ResponseWriter, r *http.Request) {
	var input application.UpdatePlanInput
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
	plan, err := h.service.UpdatePlan(r.Context(), sharedhttp.PathValue(r, "tenantID"), sharedhttp.PathValue(r, "planID"), input)
	if err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	sharedhttp.WriteOK(w, plan)
}

func (h *Handler) ListPlans(w http.ResponseWriter, r *http.Request) {
	query, err := shareddomain.ParsePageQuery(r.URL.Query())
	if err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	items, total, err := h.service.ListPlans(r.Context(), sharedhttp.PathValue(r, "tenantID"), query)
	if err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	sharedhttp.WriteJSON(w, http.StatusOK, sharedhttp.Envelope{Data: items, Meta: shareddomain.NewPage(toAnyPlan(items), query, total)})
}

func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request) {
	task, err := h.service.GetTask(r.Context(), sharedhttp.PathValue(r, "tenantID"), sharedhttp.PathValue(r, "taskID"))
	if err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	sharedhttp.WriteOK(w, task)
}

func (h *Handler) TransitionTask(w http.ResponseWriter, r *http.Request) {
	var input application.TransitionTaskInput
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
	task, err := h.service.TransitionTask(r.Context(), sharedhttp.PathValue(r, "tenantID"), sharedhttp.PathValue(r, "taskID"), input)
	if err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	sharedhttp.WriteOK(w, task)
}

func (h *Handler) ListTasks(w http.ResponseWriter, r *http.Request) {
	query, err := shareddomain.ParsePageQuery(r.URL.Query())
	if err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	items, total, err := h.service.ListTasks(r.Context(), sharedhttp.PathValue(r, "tenantID"), query)
	if err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	sharedhttp.WriteJSON(w, http.StatusOK, sharedhttp.Envelope{Data: items, Meta: shareddomain.NewPage(toAnyTask(items), query, total)})
}

func toAnyPlan(items []maintenancedomain.Plan) []any {
	out := make([]any, len(items))
	for i := range items {
		out[i] = items[i]
	}
	return out
}

func toAnyTask(items []maintenancedomain.Task) []any {
	out := make([]any, len(items))
	for i := range items {
		out[i] = items[i]
	}
	return out
}
