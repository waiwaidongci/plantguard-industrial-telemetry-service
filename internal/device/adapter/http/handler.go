package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/acme/plantguard/internal/device/application"
	devicedomain "github.com/acme/plantguard/internal/device/domain"
	sharedhttp "github.com/acme/plantguard/internal/shared/adapter/http"
	shareddomain "github.com/acme/plantguard/internal/shared/domain"
)

type Handler struct {
	service application.Service
}

func NewHandler(service application.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateModel(w http.ResponseWriter, r *http.Request) {
	var input application.CreateModelInput
	if err := sharedhttp.DecodeJSON(r, &input); err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	model, created, err := h.service.CreateModel(r.Context(), sharedhttp.PathValue(r, "tenantID"), r.Header.Get("Idempotency-Key"), input)
	if err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	status := http.StatusCreated
	if !created {
		status = http.StatusOK
	}
	sharedhttp.WriteJSON(w, status, sharedhttp.Envelope{Data: model})
}

func (h *Handler) GetModel(w http.ResponseWriter, r *http.Request) {
	model, err := h.service.GetModel(r.Context(), sharedhttp.PathValue(r, "tenantID"), sharedhttp.PathValue(r, "modelID"))
	if err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	sharedhttp.WriteOK(w, model)
}

func (h *Handler) UpdateModel(w http.ResponseWriter, r *http.Request) {
	var input application.UpdateModelInput
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
	model, err := h.service.UpdateModel(r.Context(), sharedhttp.PathValue(r, "tenantID"), sharedhttp.PathValue(r, "modelID"), input)
	if err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	sharedhttp.WriteOK(w, model)
}

func (h *Handler) ListModels(w http.ResponseWriter, r *http.Request) {
	query, err := shareddomain.ParsePageQuery(r.URL.Query())
	if err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	items, total, err := h.service.ListModels(r.Context(), sharedhttp.PathValue(r, "tenantID"), query)
	if err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	sharedhttp.WriteJSON(w, http.StatusOK, sharedhttp.Envelope{Data: items, Meta: shareddomain.NewPage(toAnyModel(items), query, total)})
}

func (h *Handler) CreateDevice(w http.ResponseWriter, r *http.Request) {
	var input application.CreateDeviceInput
	if err := sharedhttp.DecodeJSON(r, &input); err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	device, created, err := h.service.CreateDevice(r.Context(), sharedhttp.PathValue(r, "tenantID"), r.Header.Get("Idempotency-Key"), input)
	if err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	status := http.StatusCreated
	if !created {
		status = http.StatusOK
	}
	sharedhttp.WriteJSON(w, status, sharedhttp.Envelope{Data: device})
}

func (h *Handler) GetDevice(w http.ResponseWriter, r *http.Request) {
	device, err := h.service.GetDevice(r.Context(), sharedhttp.PathValue(r, "tenantID"), sharedhttp.PathValue(r, "deviceID"))
	if err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	sharedhttp.WriteOK(w, device)
}

func (h *Handler) UpdateDevice(w http.ResponseWriter, r *http.Request) {
	var input application.UpdateDeviceInput
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
	device, err := h.service.UpdateDevice(r.Context(), sharedhttp.PathValue(r, "tenantID"), sharedhttp.PathValue(r, "deviceID"), input)
	if err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	sharedhttp.WriteOK(w, device)
}

func (h *Handler) ListDevices(w http.ResponseWriter, r *http.Request) {
	query, err := shareddomain.ParsePageQuery(r.URL.Query())
	if err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	items, total, err := h.service.ListDevices(r.Context(), sharedhttp.PathValue(r, "tenantID"), query)
	if err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	sharedhttp.WriteJSON(w, http.StatusOK, sharedhttp.Envelope{Data: items, Meta: shareddomain.NewPage(toAnyDevice(items), query, total)})
}

func (h *Handler) Heartbeat(w http.ResponseWriter, r *http.Request) {
	var input application.HeartbeatInput
	if r.ContentLength > 0 {
		if err := sharedhttp.DecodeJSON(r, &input); err != nil {
			sharedhttp.WriteError(w, r, err)
			return
		}
	}
	if input.ObservedAt.IsZero() {
		input.ObservedAt = time.Now().UTC()
	}
	device, err := h.service.Heartbeat(r.Context(), sharedhttp.PathValue(r, "tenantID"), sharedhttp.PathValue(r, "deviceID"), input)
	if err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	sharedhttp.WriteOK(w, device)
}

func toAnyModel(items []devicedomain.DeviceModel) []any {
	out := make([]any, len(items))
	for i := range items {
		out[i] = items[i]
	}
	return out
}

func toAnyDevice(items []devicedomain.Device) []any {
	out := make([]any, len(items))
	for i := range items {
		out[i] = items[i]
	}
	return out
}
