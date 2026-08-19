package http

import (
	"net/http"
	"strconv"

	"github.com/acme/plantguard/internal/event/application"
	eventdomain "github.com/acme/plantguard/internal/event/domain"
	sharedhttp "github.com/acme/plantguard/internal/shared/adapter/http"
	shareddomain "github.com/acme/plantguard/internal/shared/domain"
)

type Handler struct {
	service application.Service
}

func NewHandler(service application.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var input application.CreateEventInput
	if err := sharedhttp.DecodeJSON(r, &input); err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	event, created, err := h.service.Create(r.Context(), sharedhttp.PathValue(r, "tenantID"), r.Header.Get("Idempotency-Key"), input)
	if err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	status := http.StatusCreated
	if !created {
		status = http.StatusOK
	}
	sharedhttp.WriteJSON(w, status, sharedhttp.Envelope{Data: event})
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	event, err := h.service.Get(r.Context(), sharedhttp.PathValue(r, "tenantID"), sharedhttp.PathValue(r, "eventID"))
	if err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	sharedhttp.WriteOK(w, event)
}

func (h *Handler) Acknowledge(w http.ResponseWriter, r *http.Request) {
	var input application.AcknowledgeEventInput
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
	event, err := h.service.Acknowledge(r.Context(), sharedhttp.PathValue(r, "tenantID"), sharedhttp.PathValue(r, "eventID"), input)
	if err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	sharedhttp.WriteOK(w, event)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	query, err := shareddomain.ParsePageQuery(r.URL.Query())
	if err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	items, total, err := h.service.List(r.Context(), sharedhttp.PathValue(r, "tenantID"), query)
	if err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	sharedhttp.WriteJSON(w, http.StatusOK, sharedhttp.Envelope{Data: items, Meta: shareddomain.NewPage(toAnyEvent(items), query, total)})
}

func toAnyEvent(items []eventdomain.Event) []any {
	out := make([]any, len(items))
	for i := range items {
		out[i] = items[i]
	}
	return out
}
