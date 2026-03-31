package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/clubepay/backend/internal/domain"
	"github.com/clubepay/backend/internal/middleware"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- GetSpendingStatus tests ----

func TestGetSpendingStatus_Success(t *testing.T) {
	h := setupHandler(t)

	// Register owner via handler
	regBody := map[string]string{
		"email":         "spending@example.com",
		"password":      "password123",
		"name":          "Spending Owner",
		"business_name": "Spending Business",
		"segment":       "cafeteria",
	}
	b, _ := json.Marshal(regBody)
	regReq := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(b))
	regReq.Header.Set("Content-Type", "application/json")
	regRr := httptest.NewRecorder()
	h.RegisterOwner(regRr, regReq)
	require.Equal(t, http.StatusCreated, regRr.Code)

	var regResp map[string]interface{}
	require.NoError(t, json.NewDecoder(regRr.Body).Decode(&regResp))
	userMap := regResp["user"].(map[string]interface{})
	userID := int64(userMap["id"].(float64))

	// Create monthly cost entry
	ctx := context.Background()
	business, err := h.Queries.GetBusinessByOwnerID(ctx, userID)
	require.NoError(t, err)

	currentMonth := domain.GetCurrentMonth()
	monthlyCost, err := h.Queries.CreateMonthlyCost(ctx, map[string]interface{}{
		"business_id":          business.ID,
		"month":                pgTypeDate(currentMonth),
		"monthly_budget_cents": int64(500000), // $5000
	})
	require.NoError(t, err)

	// Make request with auth
	req := httptest.NewRequest(http.MethodGet, "/api/owner/spending/status", nil)
	req = withAuth(req, userID, "owner")
	rr := httptest.NewRecorder()

	h.GetSpendingStatus(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var respBody map[string]interface{}
	err = json.NewDecoder(rr.Body).Decode(&respBody)
	require.NoError(t, err)

	assert.Equal(t, float64(monthlyCost.TotalCostCents), respBody["current_cost_cents"])
	assert.Equal(t, float64(monthlyCost.MonthlyBudgetCents), respBody["budget_cents"])
}

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

func TestGetSpendingHistory_Success(t *testing.T) {
	h := setupHandler(t)

	// Register owner
	regBody := map[string]string{
		"email":         "history@example.com",
		"password":      "password123",
		"name":          "History Owner",
		"business_name": "History Business",
		"segment":       "cafeteria",
	}
	b, _ := json.Marshal(regBody)
	regReq := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(b))
	regReq.Header.Set("Content-Type", "application/json")
	regRr := httptest.NewRecorder()
	h.RegisterOwner(regRr, regReq)
	require.Equal(t, http.StatusCreated, regRr.Code)

	var regResp map[string]interface{}
	require.NoError(t, json.NewDecoder(regRr.Body).Decode(&regResp))
	userMap := regResp["user"].(map[string]interface{})
	userID := int64(userMap["id"].(float64))

	req := httptest.NewRequest(http.MethodGet, "/api/owner/spending/history?limit=10&offset=0", nil)
	req = withAuth(req, userID, "owner")
	rr := httptest.NewRecorder()

	h.GetSpendingHistory(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var respBody map[string]interface{}
	err := json.NewDecoder(rr.Body).Decode(&respBody)
	require.NoError(t, err)

	assert.IsType(t, []interface{}{}, respBody["items"])
	assert.Equal(t, float64(10), respBody["limit"])
	assert.Equal(t, float64(0), respBody["offset"])
}

func TestGetSpendingHistory_Unauthorized(t *testing.T) {
	h := setupHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/owner/spending/history", nil)
	rr := httptest.NewRecorder()

	h.GetSpendingHistory(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

// ---- GetAlertHistory tests ----

func TestGetAlertHistory_Success(t *testing.T) {
	h := setupHandler(t)

	// Register owner
	regBody := map[string]string{
		"email":         "alerts@example.com",
		"password":      "password123",
		"name":          "Alerts Owner",
		"business_name": "Alerts Business",
		"segment":       "cafeteria",
	}
	b, _ := json.Marshal(regBody)
	regReq := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(b))
	regReq.Header.Set("Content-Type", "application/json")
	regRr := httptest.NewRecorder()
	h.RegisterOwner(regRr, regReq)
	require.Equal(t, http.StatusCreated, regRr.Code)

	var regResp map[string]interface{}
	require.NoError(t, json.NewDecoder(regRr.Body).Decode(&regResp))
	userMap := regResp["user"].(map[string]interface{})
	userID := int64(userMap["id"].(float64))

	req := httptest.NewRequest(http.MethodGet, "/api/owner/spending/alerts?limit=10&offset=0", nil)
	req = withAuth(req, userID, "owner")
	rr := httptest.NewRecorder()

	h.GetAlertHistory(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var respBody map[string]interface{}
	err := json.NewDecoder(rr.Body).Decode(&respBody)
	require.NoError(t, err)

	assert.IsType(t, []interface{}{}, respBody["items"])
	assert.Equal(t, float64(10), respBody["limit"])
	assert.Equal(t, float64(0), respBody["offset"])
}

func TestGetAlertHistory_Unauthorized(t *testing.T) {
	h := setupHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/owner/spending/alerts", nil)
	rr := httptest.NewRecorder()

	h.GetAlertHistory(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

// Helper function to convert time.Time to pgtype.Date
func pgTypeDate(t time.Time) pgtype.Date {
	return pgtype.Date{
		Time:  t,
		Valid: true,
	}
}
