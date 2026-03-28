package middleware_test

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/clubepay/backend/internal/middleware"
	"github.com/stretchr/testify/assert"
)

func TestLogging_LogsRequestDetails(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	slog.SetDefault(logger)

	handler := middleware.Logging(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	logOutput := buf.String()
	assert.True(t, strings.Contains(logOutput, "GET"), "log should contain HTTP method")
	assert.True(t, strings.Contains(logOutput, "/api/health"), "log should contain path")
	assert.True(t, strings.Contains(logOutput, "200"), "log should contain status code")
}

func TestLogging_CapturesStatusCode(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	slog.SetDefault(logger)

	handler := middleware.Logging(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)

	logOutput := buf.String()
	assert.True(t, strings.Contains(logOutput, "404"), "log should contain 404 status")
}
