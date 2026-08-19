package http

import (
	"encoding/json"
	"errors"
	"net/http"

	shareddomain "github.com/acme/plantguard/internal/shared/domain"
)

type Envelope struct {
	Data any `json:"data,omitempty"`
	Meta any `json:"meta,omitempty"`
}

type ErrorEnvelope struct {
	Code      string         `json:"code"`
	Message   string         `json:"message"`
	RequestID string         `json:"request_id"`
	Details   map[string]any `json:"details,omitempty"`
}

func WriteJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if body == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(body); err != nil {
		// The header is already written; logging is the only safe action here.
		_ = err
	}
}

func WriteCreated(w http.ResponseWriter, body any) {
	WriteJSON(w, http.StatusCreated, Envelope{Data: body})
}

func WriteOK(w http.ResponseWriter, body any) {
	WriteJSON(w, http.StatusOK, Envelope{Data: body})
}

func WriteNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	var appErr *shareddomain.Error
	if !errors.As(err, &appErr) {
		appErr = shareddomain.ErrInternal
	}
	status := appErr.Status
	if status == 0 {
		status = http.StatusInternalServerError
	}
	requestID := RequestIDFromContext(r.Context())
	if appErr.RequestID != "" {
		requestID = appErr.RequestID
	}
	WriteJSON(w, status, ErrorEnvelope{
		Code:      appErr.Code,
		Message:   appErr.Message,
		RequestID: requestID,
		Details:   appErr.Details,
	})
}
