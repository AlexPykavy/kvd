package middleware

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	HttpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "route", "status"},
	)
	HttpRequestsDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_milliseconds",
			Help:    "HTTP request duration in milliseconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "route"},
	)
)

func init() {
	prometheus.MustRegister(HttpRequestsTotal)
	prometheus.MustRegister(HttpRequestsDuration)
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(b)
}

func InstrumentMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			sw := &statusWriter{
				ResponseWriter: w,
			}

			next.ServeHTTP(sw, r)

			status := sw.status
			if status == 0 {
				status = http.StatusOK
			}

			// Ideally use the router's normalized route pattern here,
			// e.g. "/users/:id", rather than r.URL.Path.
			route := r.Pattern
			if route == "" {
				route = r.URL.Path
			}

			duration := time.Since(start)

			// metrics
			HttpRequestsTotal.WithLabelValues(
				r.Method,
				route,
				strconv.Itoa(status),
			).Inc()

			HttpRequestsDuration.WithLabelValues(
				r.Method,
				route,
			).Observe(duration.Seconds() * 1000)

			// logs
			logger.Info("http request",
				"method", r.Method,
				"url", r.URL.Path,
				"status", status,
				"duration_μs", duration.Microseconds(),
			)
		})
	}
}
