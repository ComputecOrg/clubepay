package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- GetPublicBusiness tests ----

func TestGetPublicBusiness_Success(t *testing.T) {
	h := setupHandler(t)

	// Register owner (which creates business with a slug) via handler
	regBody := map[string]interface{}{
		"email":         "pubowner@example.com",
		"password":      "password123",
		"name":          "Pub Owner",
		"business_name": "Café Público",
		"segment":       "cafeteria",
	}
	rb, _ := json.Marshal(regBody)
	regReq := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(rb))
	regReq.Header.Set("Content-Type", "application/json")
	regRr := httptest.NewRecorder()
	h.RegisterOwner(regRr, regReq)
	require.Equal(t, http.StatusCreated, regRr.Code)

	var regResp map[string]interface{}
	require.NoError(t, json.NewDecoder(regRr.Body).Decode(&regResp))
	bizMap := regResp["business"].(map[string]interface{})
	slug := bizMap["slug"].(string)

	// GET public business by slug
	req := httptest.NewRequest(http.MethodGet, "/api/public/business/"+slug, nil)
	r := chi.NewRouter()
	r.Get("/api/public/business/{slug}", h.GetPublicBusiness)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, "Café Público", resp["name"])
	assert.Equal(t, slug, resp["slug"])
	assert.Contains(t, resp, "subscriber_count")
}

func TestGetPublicBusiness_NotFound(t *testing.T) {
	h := setupHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/public/business/nonexistent-slug-xyz", nil)
	r := chi.NewRouter()
	r.Get("/api/public/business/{slug}", h.GetPublicBusiness)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// ---- GetPublicPlans tests ----

func TestGetPublicPlans_Success(t *testing.T) {
	h := setupHandler(t)

	// Register owner (which creates business) via handler
	regBody := map[string]interface{}{
		"email":         "pubplans@example.com",
		"password":      "password123",
		"name":          "Pub Plans Owner",
		"business_name": "Café Planos Público",
		"segment":       "cafeteria",
	}
	rb, _ := json.Marshal(regBody)
	regReq := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(rb))
	regReq.Header.Set("Content-Type", "application/json")
	regRr := httptest.NewRecorder()
	h.RegisterOwner(regRr, regReq)
	require.Equal(t, http.StatusCreated, regRr.Code)

	var regResp map[string]interface{}
	require.NoError(t, json.NewDecoder(regRr.Body).Decode(&regResp))
	userMap := regResp["user"].(map[string]interface{})
	bizMap := regResp["business"].(map[string]interface{})
	userID := int64(userMap["id"].(float64))
	slug := bizMap["slug"].(string)

	// Create a plan
	planBody := map[string]interface{}{
		"name":        "Plano Público",
		"price_cents": 2990,
		"limit_type":  "daily",
		"limit_count": 1,
	}
	pb, _ := json.Marshal(planBody)
	planReq := httptest.NewRequest(http.MethodPost, "/api/plans", bytes.NewReader(pb))
	planReq.Header.Set("Content-Type", "application/json")
	planReq = withAuth(planReq, userID, "owner")
	planRr := httptest.NewRecorder()
	h.CreatePlan(planRr, planReq)
	require.Equal(t, http.StatusCreated, planRr.Code)

	// GET public plans by slug
	req := httptest.NewRequest(http.MethodGet, "/api/public/plans/"+slug, nil)
	r := chi.NewRouter()
	r.Get("/api/public/plans/{slug}", h.GetPublicPlans)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp []interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Len(t, resp, 1)
}
