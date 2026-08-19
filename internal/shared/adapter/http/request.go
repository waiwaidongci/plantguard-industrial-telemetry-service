package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	shareddomain "github.com/acme/plantguard/internal/shared/domain"
)

type contextKey string

const requestIDKey contextKey = "request_id"

func RequestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey).(string)
	return id
}

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

func DecodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return shareddomain.New("invalid_json", "request body is not valid JSON or contains unknown fields", http.StatusBadRequest)
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return shareddomain.New("invalid_json", "request body must contain a single JSON object", http.StatusBadRequest)
	}
	return nil
}

func PathValue(r *http.Request, name string) string {
	return r.PathValue(name)
}
