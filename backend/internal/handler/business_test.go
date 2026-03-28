package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- GetBusiness tests ----

func TestGetBusiness_Success(t *testing.T) {
	h := setupHandler(t)

	// Register owner via handler (same DB as h) to ensure data is in the right DB
	regBody := map[string]string{
		"email":         "getbiz@example.com",
		"password":      "password123",
		"name":          "Get Biz Owner",
		"business_name": "Get Biz Café",
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

	req := httptest.NewRequest(http.MethodGet, "/api/business", nil)
	req = withAuth(req, userID, "owner")
	rr := httptest.NewRecorder()

	h.GetBusiness(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, "Get Biz Café", resp["name"])
	assert.NotEmpty(t, resp["slug"])
}

func TestGetBusiness_NotFound(t *testing.T) {
	h := setupHandler(t)

	// Use a user ID that doesn't have a business
	req := httptest.NewRequest(http.MethodGet, "/api/business", nil)
	req = withAuth(req, 999999, "owner")
	rr := httptest.NewRecorder()

	h.GetBusiness(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// ---- UpdateBusiness tests ----

func TestUpdateBusiness_Success(t *testing.T) {
	h := setupHandler(t)

	// Register owner via handler (same DB as h)
	regBody := map[string]string{
		"email":         "updatebiz@example.com",
		"password":      "password123",
		"name":          "Update Biz Owner",
		"business_name": "Update Biz Original",
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

	updateBody := map[string]string{
		"name":     "Update Biz Novo Nome",
		"segment":  "padaria",
		"address":  "Rua Teste 123",
		"logo_url": "https://example.com/logo.png",
	}
	ub, _ := json.Marshal(updateBody)
	req := httptest.NewRequest(http.MethodPut, "/api/business", bytes.NewReader(ub))
	req.Header.Set("Content-Type", "application/json")
	req = withAuth(req, userID, "owner")
	rr := httptest.NewRecorder()

	h.UpdateBusiness(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, "Update Biz Novo Nome", resp["name"])
	assert.Equal(t, "padaria", resp["segment"])
}
