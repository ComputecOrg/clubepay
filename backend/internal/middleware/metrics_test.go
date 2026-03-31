package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/clubepay/backend/internal/middleware"
)

func TestPrometheusMiddleware_RecordsRequest(t *testing.T) {
	reg := prometheus.NewRegistry()
	mw := middleware.NewPrometheus(reg)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/plans", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	mfs, err := reg.Gather()
	require.NoError(t, err)
	assert.Greater(t, len(mfs), 0, "deve haver ao menos uma métrica registrada")

	var found bool
	for _, mf := range mfs {
		if mf.GetName() == "http_requests_total" {
			found = true
			break
		}
	}
	assert.True(t, found, "métrica http_requests_total deve existir")
}

func TestPrometheusMiddleware_RecordsStatusCode(t *testing.T) {
	reg := prometheus.NewRegistry()
	mw := middleware.NewPrometheus(reg)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/unknown", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	mfs, err := reg.Gather()
	require.NoError(t, err)

	for _, mf := range mfs {
		if mf.GetName() == "http_requests_total" {
			for _, m := range mf.GetMetric() {
				for _, label := range m.GetLabel() {
					if label.GetName() == "status" {
						assert.Equal(t, "404", label.GetValue())
					}
				}
			}
		}
	}
}
