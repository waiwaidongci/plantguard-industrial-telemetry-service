package http

import (
	"net/http"
	"time"

	sharedhttp "github.com/acme/plantguard/internal/shared/adapter/http"
	shareddomain "github.com/acme/plantguard/internal/shared/domain"
	"github.com/acme/plantguard/internal/telemetry/application"
	telemetrydomain "github.com/acme/plantguard/internal/telemetry/domain"
)

type Handler struct {
	service application.Service
}

func NewHandler(service application.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) IngestBatch(w http.ResponseWriter, r *http.Request) {
	var input application.IngestBatchInput
	if err := sharedhttp.DecodeJSON(r, &input); err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	batch, err := h.service.IngestBatch(r.Context(), sharedhttp.PathValue(r, "tenantID"), sharedhttp.PathValue(r, "deviceID"), input)
	if err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	sharedhttp.WriteJSON(w, http.StatusCreated, sharedhttp.Envelope{Data: batch})
}

func (h *Handler) Summaries(w http.ResponseWriter, r *http.Request) {
	since, until, err := parseRange(r)
	if err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	items, err := h.service.Summaries(r.Context(), sharedhttp.PathValue(r, "tenantID"), sharedhttp.PathValue(r, "deviceID"), application.SummaryQuery{Since: since, Until: until})
	if err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	sharedhttp.WriteOK(w, items)
}

func (h *Handler) ListReadings(w http.ResponseWriter, r *http.Request) {
	query, err := shareddomain.ParsePageQuery(r.URL.Query())
	if err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	items, total, err := h.service.ListReadings(r.Context(), sharedhttp.PathValue(r, "tenantID"), sharedhttp.PathValue(r, "deviceID"), query)
	if err != nil {
		sharedhttp.WriteError(w, r, err)
		return
	}
	sharedhttp.WriteJSON(w, http.StatusOK, sharedhttp.Envelope{Data: items, Meta: shareddomain.NewPage(toAnyReading(items), query, total)})
}

func parseRange(r *http.Request) (time.Time, time.Time, error) {
	since := time.Time{}
	until := time.Time{}
	if raw := r.URL.Query().Get("since"); raw != "" {
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return since, until, shareddomain.New("invalid_since", "since must be an RFC3339 timestamp", 400)
		}
		since = t
	}
	if raw := r.URL.Query().Get("until"); raw != "" {
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return since, until, shareddomain.New("invalid_until", "until must be an RFC3339 timestamp", 400)
		}
		until = t
	}
	return since, until, nil
}

func toAnyReading(items []telemetrydomain.ReadingRecord) []any {
	out := make([]any, len(items))
	for i := range items {
		out[i] = items[i]
	}
	return out
}
