package http

import (
	"net/http"
	"strconv"
)

func RequestMetrics(metrics *Metrics) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			metrics.Inc("plantguard_http_requests_total")
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)
			metrics.IncLabel("plantguard_http_responses_total", strconv.Itoa(rec.status))
		})
	}
}
