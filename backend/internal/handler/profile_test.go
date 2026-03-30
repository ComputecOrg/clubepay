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

// registerOwnerForProfile is a helper that registers an owner and returns their userID.
func registerOwnerForProfile(t *testing.T, h interface{ RegisterOwner(http.ResponseWriter, *http.Request) }, email string) int64 {
	t.Helper()
	regBody := map[string]string{
		"email":         email,
		"password":      "password123",
		"name":          "Test User",
		"business_name": "Test Cafe",
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
	userID := int64(regResp["user"].(map[string]interface{})["id"].(float64))
	return userID
}

// ---- GetProfile tests ----

func TestGetProfile_Success(t *testing.T) {
	h := setupHandler(t)
	userID := registerOwnerForProfile(t, h, "getprofile@example.com")

	req := httptest.NewRequest(http.MethodGet, "/api/profile", nil)
	req = withAuth(req, userID, "owner")
	rr := httptest.NewRecorder()

	h.GetProfile(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))

	user, ok := resp["user"].(map[string]interface{})
	require.True(t, ok, "response must have a 'user' field")
	assert.Equal(t, "getprofile@example.com", user["email"])
	assert.Equal(t, "Test User", user["name"])
	assert.Equal(t, "owner", user["role"])
	assert.NotEmpty(t, user["id"])
}

// ---- UpdateProfile tests ----

func TestUpdateProfile_Success(t *testing.T) {
	h := setupHandler(t)
	userID := registerOwnerForProfile(t, h, "updateprofile@example.com")

	updateBody := map[string]string{
		"name":  "Updated Name",
		"phone": "11988887777",
	}
	b, _ := json.Marshal(updateBody)
	req := httptest.NewRequest(http.MethodPut, "/api/profile", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req = withAuth(req, userID, "owner")
	rr := httptest.NewRecorder()

	h.UpdateProfile(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))

	user, ok := resp["user"].(map[string]interface{})
	require.True(t, ok, "response must have a 'user' field")
	assert.Equal(t, "Updated Name", user["name"])
	assert.Equal(t, "11988887777", user["phone"])
}

func TestUpdateProfile_InvalidJSON(t *testing.T) {
	h := setupHandler(t)
	userID := registerOwnerForProfile(t, h, "updateprofile-badjson@example.com")

	req := httptest.NewRequest(http.MethodPut, "/api/profile", bytes.NewReader([]byte(`{invalid json`)))
	req.Header.Set("Content-Type", "application/json")
	req = withAuth(req, userID, "owner")
	rr := httptest.NewRecorder()

	h.UpdateProfile(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestUpdateProfile_MissingName(t *testing.T) {
	h := setupHandler(t)
	userID := registerOwnerForProfile(t, h, "updateprofile-noname@example.com")

	updateBody := map[string]string{
		"phone": "11988887777",
	}
	b, _ := json.Marshal(updateBody)
	req := httptest.NewRequest(http.MethodPut, "/api/profile", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req = withAuth(req, userID, "owner")
	rr := httptest.NewRecorder()

	h.UpdateProfile(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// ---- ChangePassword tests ----

func TestChangePassword_Success(t *testing.T) {
	h := setupHandler(t)
	userID := registerOwnerForProfile(t, h, "changepw@example.com")

	body := map[string]string{
		"current_password": "password123",
		"new_password":     "newpassword456",
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/profile/change-password", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req = withAuth(req, userID, "owner")
	rr := httptest.NewRecorder()

	h.ChangePassword(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Contains(t, resp["message"], "senha")
}

func TestChangePassword_WrongCurrent(t *testing.T) {
	h := setupHandler(t)
	userID := registerOwnerForProfile(t, h, "changepw-wrong@example.com")

	body := map[string]string{
		"current_password": "wrongpassword",
		"new_password":     "newpassword456",
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/profile/change-password", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req = withAuth(req, userID, "owner")
	rr := httptest.NewRecorder()

	h.ChangePassword(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestChangePassword_ShortPassword(t *testing.T) {
	h := setupHandler(t)
	userID := registerOwnerForProfile(t, h, "changepw-short@example.com")

	body := map[string]string{
		"current_password": "password123",
		"new_password":     "short",
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/profile/change-password", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req = withAuth(req, userID, "owner")
	rr := httptest.NewRecorder()

	h.ChangePassword(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestChangePassword_InvalidJSON(t *testing.T) {
	h := setupHandler(t)
	userID := registerOwnerForProfile(t, h, "changepw-badjson@example.com")

	req := httptest.NewRequest(http.MethodPost, "/api/profile/change-password", bytes.NewReader([]byte(`{invalid json`)))
	req.Header.Set("Content-Type", "application/json")
	req = withAuth(req, userID, "owner")
	rr := httptest.NewRecorder()

	h.ChangePassword(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
