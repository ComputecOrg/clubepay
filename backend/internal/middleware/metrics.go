package middleware

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
)

// responseWriter captura o status code da resposta HTTP.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// NewPrometheus retorna um middleware que registra métricas HTTP no registry fornecido.
func NewPrometheus(reg prometheus.Registerer) func(http.Handler) http.Handler {
	requestsTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Número total de requisições HTTP recebidas.",
	}, []string{"method", "path", "status"})

	requestDuration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "Duração das requisições HTTP em segundos.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path"})

	reg.MustRegister(requestsTotal, requestDuration)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			timer := prometheus.NewTimer(requestDuration.WithLabelValues(r.Method, r.URL.Path))
			defer timer.ObserveDuration()

			wrapped := newResponseWriter(w)
			next.ServeHTTP(wrapped, r)

			requestsTotal.WithLabelValues(
				r.Method,
				r.URL.Path,
				fmt.Sprintf("%d", wrapped.statusCode),
			).Inc()
		})
	}
}
