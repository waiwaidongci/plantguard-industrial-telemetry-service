package http

import (
	"net/http"
	"strconv"

	"github.com/acme/plantguard/internal/rule/application"
	ruledomain "github.com/acme/plantguard/internal/rule/domain"
	sharedhttp "github.com/acme/plantguard/internal/shared/adapter/http"
	shareddomain "github.com/acme/plantguard/internal/shared/domain"
)

type Handler struct {
	service application.Service
}

func NewHandler(service application.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateRule(w http.ResponseWriter, r *http.Request) {
	var input application.CreateRuleInput
	if err := sharedhttp.DecodeJSON(r, &input); err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	rule, created, err := h.service.CreateRule(r.Context(), sharedhttp.PathValue(r, "tenantID"), r.Header.Get("Idempotency-Key"), input)
	if err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	status := http.StatusCreated
	if !created {
		status = http.StatusOK
	}
	sharedhttp.WriteJSON(w, status, sharedhttp.Envelope{Data: rule})
}

func (h *Handler) GetRule(w http.ResponseWriter, r *http.Request) {
	rule, err := h.service.GetRule(r.Context(), sharedhttp.PathValue(r, "tenantID"), sharedhttp.PathValue(r, "ruleID"))
	if err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	sharedhttp.WriteOK(w, rule)
}

func (h *Handler) UpdateRule(w http.ResponseWriter, r *http.Request) {
	var input application.UpdateRuleInput
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
	rule, err := h.service.UpdateRule(r.Context(), sharedhttp.PathValue(r, "tenantID"), sharedhttp.PathValue(r, "ruleID"), input)
	if err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	sharedhttp.WriteOK(w, rule)
}

func (h *Handler) ListRules(w http.ResponseWriter, r *http.Request) {
	query, err := shareddomain.ParsePageQuery(r.URL.Query())
	if err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	items, total, err := h.service.ListRules(r.Context(), sharedhttp.PathValue(r, "tenantID"), query)
	if err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	sharedhttp.WriteJSON(w, http.StatusOK, sharedhttp.Envelope{Data: items, Meta: shareddomain.NewPage(toAnyRule(items), query, total)})
}

func toAnyRule(items []ruledomain.Rule) []any {
	out := make([]any, len(items))
	for i := range items {
		out[i] = items[i]
	}
	return out
}
