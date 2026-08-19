package http

import (
	"net/http"

	"github.com/acme/plantguard/internal/notification/application"
	notificationdomain "github.com/acme/plantguard/internal/notification/domain"
	sharedhttp "github.com/acme/plantguard/internal/shared/adapter/http"
	shareddomain "github.com/acme/plantguard/internal/shared/domain"
)

type Handler struct {
	service application.Service
}

func NewHandler(service application.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Send(w http.ResponseWriter, r *http.Request) {
	var input application.SendNotificationInput
	if err := sharedhttp.DecodeJSON(r, &input); err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	notification, err := h.service.Send(r.Context(), sharedhttp.PathValue(r, "tenantID"), input)
	if err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	sharedhttp.WriteJSON(w, http.StatusCreated, sharedhttp.Envelope{Data: notification})
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
	sharedhttp.WriteJSON(w, http.StatusOK, sharedhttp.Envelope{Data: items, Meta: shareddomain.NewPage(toAnyNotification(items), query, total)})
}

func toAnyNotification(items []notificationdomain.Notification) []any {
	out := make([]any, len(items))
	for i := range items {
		out[i] = items[i]
	}
	return out
}
