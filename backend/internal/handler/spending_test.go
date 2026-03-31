package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ---- GetSpendingStatus tests ----

func TestGetSpendingStatus_Unauthorized(t *testing.T) {
	h := setupHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/owner/spending/status", nil)
	rr := httptest.NewRecorder()

	h.GetSpendingStatus(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestGetSpendingStatus_BusinessNotFound(t *testing.T) {
	h := setupHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/owner/spending/status", nil)
	req = withAuth(req, 999999, "owner")
	rr := httptest.NewRecorder()

	h.GetSpendingStatus(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// ---- GetSpendingHistory tests ----

func TestGetSpendingHistory_Unauthorized(t *testing.T) {
	h := setupHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/owner/spending/history", nil)
	rr := httptest.NewRecorder()

	h.GetSpendingHistory(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

// ---- GetAlertHistory tests ----

func TestGetAlertHistory_Unauthorized(t *testing.T) {
	h := setupHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/owner/spending/alerts", nil)
	rr := httptest.NewRecorder()

	h.GetAlertHistory(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}
