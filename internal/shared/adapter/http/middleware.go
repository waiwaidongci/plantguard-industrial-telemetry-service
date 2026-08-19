package http

import (
	"context"
	"log/slog"
	"net/http"
	"runtime/debug"
	"sync"
	"time"

	shareddomain "github.com/acme/plantguard/internal/shared/domain"
)

type Middleware func(http.Handler) http.Handler

func Chain(handler http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	n, err := r.ResponseWriter.Write(b)
	r.bytes += n
	return n, err
}

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = shareddomain.NewID("req")
		}
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(WithRequestID(r.Context(), id)))
	})
}

func Timeout(timeout time.Duration) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func AccessLog(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)
			logger.Info("http_request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rec.status,
				"bytes", rec.bytes,
				"duration_ms", time.Since(start).Milliseconds(),
				"request_id", RequestIDFromContext(r.Context()),
			)
		})
	}
}

func RecoverPanic(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if recovered := recover(); recovered != nil {
					logger.Error("panic_recovered",
						"panic", recovered,
						"stack", string(debug.Stack()),
						"request_id", RequestIDFromContext(r.Context()),
					)
					WriteError(w, r, shareddomain.ErrInternal)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func AuthPlaceholder(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Authentication is a placeholder. The token is available
		// to future adapters without changing route signatures.
		token := r.Header.Get("Authorization")
		if token != "" && len(token) > 128 {
			WriteError(w, r, shareddomain.ErrUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type tokenBucket struct {
	mu     sync.Mutex
	tokens float64
	last   time.Time
	rate   float64
	burst  float64
}

func newTokenBucket(rate, burst int) *tokenBucket {
	return &tokenBucket{rate: float64(rate), burst: float64(burst), tokens: float64(burst), last: time.Now()}
}

func (b *tokenBucket) allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	now := time.Now()
	elapsed := now.Sub(b.last).Seconds()
	b.tokens = min(b.burst, b.tokens+elapsed*b.rate)
	b.last = now
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

func RateLimit(rate, burst int) Middleware {
	bucket := newTokenBucket(rate, burst)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !bucket.allow() {
				WriteError(w, r, shareddomain.ErrTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
