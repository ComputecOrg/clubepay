package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/clubepay/backend/internal/middleware"
)

func TestSentryRecovery_CatchesPanic(t *testing.T) {
	// Sem DSN configurado (modo local/dev), o middleware deve apenas recuperar o panic
	mw := middleware.SentryRecovery("")

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("erro inesperado no handler")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/plans", nil)
	w := httptest.NewRecorder()

	assert.NotPanics(t, func() {
		handler.ServeHTTP(w, req)
	})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestSentryRecovery_PassthroughNormal(t *testing.T) {
	mw := middleware.SentryRecovery("")

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/plans", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
