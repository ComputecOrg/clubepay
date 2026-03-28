package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/clubepay/backend/internal/config"
	"github.com/clubepay/backend/internal/handler"
	"github.com/clubepay/backend/internal/psp"
	"github.com/clubepay/backend/internal/repository"
	"github.com/clubepay/backend/internal/testutil"
	"github.com/stretchr/testify/assert"
)

// setupBrokenHandler creates a handler with a closed DB pool.
// All DB queries will return errors, triggering internal server error branches.
func setupBrokenHandler(t *testing.T) *handler.Handler {
	t.Helper()
	pool := testutil.SetupTestDB(t)
	pool.Close() // Close pool — all queries will fail
	q := repository.New(pool)
	cfg := &config.Config{
		JWTSecret:  "test-secret",
		CronSecret: "test-cron-secret",
	}
	return handler.New(q, cfg, &psp.MockPSP{})
}

// Tests that hit internal server error branches (DB unavailable)

func TestGetBusiness_DBError(t *testing.T) {
	h := setupBrokenHandler(t)
	req := httptest.NewRequest("GET", "/api/business", nil)
	req = withAuth(req, 1, "owner")
	rr := httptest.NewRecorder()
	h.GetBusiness(rr, req)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestUpdateBusiness_DBError(t *testing.T) {
	h := setupBrokenHandler(t)
	body, _ := json.Marshal(map[string]string{"name": "test"})
	req := httptest.NewRequest("PUT", "/api/business", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withAuth(req, 1, "owner")
	rr := httptest.NewRecorder()
	h.UpdateBusiness(rr, req)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestListPlans_DBError(t *testing.T) {
	h := setupBrokenHandler(t)
	req := httptest.NewRequest("GET", "/api/plans", nil)
	req = withAuth(req, 1, "owner")
	rr := httptest.NewRecorder()
	h.ListPlans(rr, req)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestCreatePlan_DBError(t *testing.T) {
	h := setupBrokenHandler(t)
	body, _ := json.Marshal(map[string]interface{}{
		"name": "Test", "price_cents": 100, "limit_type": "daily", "limit_count": 1,
	})
	req := httptest.NewRequest("POST", "/api/plans", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withAuth(req, 1, "owner")
	rr := httptest.NewRecorder()
	h.CreatePlan(rr, req)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestUpdatePlan_DBError(t *testing.T) {
	h := setupBrokenHandler(t)
	body, _ := json.Marshal(map[string]interface{}{
		"name": "Test", "price_cents": 100, "limit_type": "daily", "limit_count": 1,
	})
	r := chi.NewRouter()
	r.Put("/api/plans/{id}", h.UpdatePlan)
	req := httptest.NewRequest("PUT", "/api/plans/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withAuth(req, 1, "owner")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestDeletePlan_DBError(t *testing.T) {
	h := setupBrokenHandler(t)
	r := chi.NewRouter()
	r.Delete("/api/plans/{id}", h.DeletePlan)
	req := httptest.NewRequest("DELETE", "/api/plans/1", nil)
	req = withAuth(req, 1, "owner")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestListSubscriptions_DBError(t *testing.T) {
	h := setupBrokenHandler(t)
	req := httptest.NewRequest("GET", "/api/subscriptions", nil)
	req = withAuth(req, 1, "owner")
	rr := httptest.NewRecorder()
	h.ListSubscriptions(rr, req)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestCancelSubscriptionByOwner_DBError(t *testing.T) {
	h := setupBrokenHandler(t)
	r := chi.NewRouter()
	r.Delete("/api/subscriptions/{id}", h.CancelSubscriptionByOwner)
	req := httptest.NewRequest("DELETE", "/api/subscriptions/1", nil)
	req = withAuth(req, 1, "owner")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestSubscribe_DBError(t *testing.T) {
	h := setupBrokenHandler(t)
	body, _ := json.Marshal(map[string]interface{}{"plan_id": 1, "cpf": "12345678900"})
	req := httptest.NewRequest("POST", "/api/subscribe", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withAuth(req, 1, "subscriber")
	rr := httptest.NewRecorder()
	h.Subscribe(rr, req)
	// Should get 404 or 500 depending on which query fails first
	assert.True(t, rr.Code >= 400)
}

func TestValidateUsage_DBError(t *testing.T) {
	h := setupBrokenHandler(t)
	body, _ := json.Marshal(map[string]string{"business_slug": "test"})
	req := httptest.NewRequest("POST", "/api/validate-usage", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withAuth(req, 1, "subscriber")
	rr := httptest.NewRecorder()
	h.ValidateUsage(rr, req)
	assert.True(t, rr.Code >= 400)
}

func TestMyUsage_DBError(t *testing.T) {
	h := setupBrokenHandler(t)
	req := httptest.NewRequest("GET", "/api/my-usage?business_slug=test", nil)
	req = withAuth(req, 1, "subscriber")
	rr := httptest.NewRecorder()
	h.MyUsage(rr, req)
	assert.True(t, rr.Code >= 400)
}

func TestMyPlan_DBError(t *testing.T) {
	h := setupBrokenHandler(t)
	req := httptest.NewRequest("GET", "/api/my-plan?business_slug=test", nil)
	req = withAuth(req, 1, "subscriber")
	rr := httptest.NewRecorder()
	h.MyPlan(rr, req)
	assert.True(t, rr.Code >= 400)
}

func TestCancelBySubscriber_DBError(t *testing.T) {
	h := setupBrokenHandler(t)
	body, _ := json.Marshal(map[string]string{"business_slug": "test"})
	req := httptest.NewRequest("POST", "/api/cancel", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withAuth(req, 1, "subscriber")
	rr := httptest.NewRecorder()
	h.CancelBySubscriber(rr, req)
	assert.True(t, rr.Code >= 400)
}

func TestMyReferralCode_DBError(t *testing.T) {
	h := setupBrokenHandler(t)
	req := httptest.NewRequest("GET", "/api/my-referral-code?business_slug=test", nil)
	req = withAuth(req, 1, "subscriber")
	rr := httptest.NewRecorder()
	h.MyReferralCode(rr, req)
	assert.True(t, rr.Code >= 400)
}

func TestApplyReferral_DBError(t *testing.T) {
	h := setupBrokenHandler(t)
	body, _ := json.Marshal(map[string]string{"code": "abc123"})
	req := httptest.NewRequest("POST", "/api/referrals/apply", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withAuth(req, 1, "subscriber")
	rr := httptest.NewRecorder()
	h.ApplyReferral(rr, req)
	assert.True(t, rr.Code >= 400)
}

func TestGetPublicBusiness_DBError(t *testing.T) {
	h := setupBrokenHandler(t)
	r := chi.NewRouter()
	r.Get("/api/public/business/{slug}", h.GetPublicBusiness)
	req := httptest.NewRequest("GET", "/api/public/business/test", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	assert.True(t, rr.Code >= 400)
}

func TestGetPublicPlans_DBError(t *testing.T) {
	h := setupBrokenHandler(t)
	r := chi.NewRouter()
	r.Get("/api/public/plans/{slug}", h.GetPublicPlans)
	req := httptest.NewRequest("GET", "/api/public/plans/test", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	assert.True(t, rr.Code >= 400)
}

func TestReconcile_DBError(t *testing.T) {
	h := setupBrokenHandler(t)
	req := httptest.NewRequest("POST", "/api/cron/reconcile", nil)
	req.Header.Set("X-Cron-Secret", "test-cron-secret")
	rr := httptest.NewRecorder()
	h.Reconcile(rr, req)
	assert.True(t, rr.Code >= 400)
}

func TestLogin_DBError(t *testing.T) {
	h := setupBrokenHandler(t)
	body, _ := json.Marshal(map[string]string{"email": "test@test.com", "password": "password123"})
	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Login(rr, req)
	assert.True(t, rr.Code >= 400)
}

func TestRegisterOwner_DBError(t *testing.T) {
	h := setupBrokenHandler(t)
	body, _ := json.Marshal(map[string]string{
		"email": "test@test.com", "password": "password123",
		"name": "Test", "business_name": "Biz",
	})
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.RegisterOwner(rr, req)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestWebhook_DBError(t *testing.T) {
	h := setupBrokenHandler(t)
	body, _ := json.Marshal(map[string]interface{}{
		"event":   "PAYMENT_CONFIRMED",
		"payment": map[string]string{"subscription": "sub_123", "status": "CONFIRMED"},
	})
	req := httptest.NewRequest("POST", "/api/psp/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.PSPWebhook(rr, req)
	// Webhook returns 200 even on unknown subscriptions (don't retry)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestRegisterSubscriber_DBError(t *testing.T) {
	h := setupBrokenHandler(t)
	body, _ := json.Marshal(map[string]string{
		"email": fmt.Sprintf("sub%d@test.com", 1), "password": "password123",
		"name": "Sub", "phone": "11999887766",
	})
	req := httptest.NewRequest("POST", "/api/auth/register-subscriber", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.RegisterSubscriber(rr, req)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}
