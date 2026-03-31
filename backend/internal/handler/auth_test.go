package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/clubepay/backend/internal/config"
	"github.com/clubepay/backend/internal/email"
	"github.com/clubepay/backend/internal/handler"
	"github.com/clubepay/backend/internal/middleware"
	"github.com/clubepay/backend/internal/psp"
	"github.com/clubepay/backend/internal/repository"
	"github.com/clubepay/backend/internal/testutil"
)

// setupHandler creates a Handler with a real testcontainers PostgreSQL DB, a mock PSP, and test config.
// SecureCookie=true simula produção.
func setupHandler(t *testing.T) *handler.Handler {
	t.Helper()
	pool := testutil.SetupTestDB(t)
	queries := repository.New(pool)
	cfg := &config.Config{
		JWTSecret:    "test-secret-key",
		SecureCookie: true,
	}
	mockPSP := &psp.MockPSP{}
	mockEmail := &email.MockSender{}
	return handler.New(queries, cfg, mockPSP, mockEmail)
}

// withAuth injects auth context (userID, role) into a request for testing authenticated endpoints.
func withAuth(r *http.Request, userID int64, role string) *http.Request {
	ctx := context.WithValue(r.Context(), middleware.UserIDContextKey(), userID)
	ctx = context.WithValue(ctx, middleware.RoleContextKey(), role)
	return r.WithContext(ctx)
}

// ---- RegisterOwner tests ----

func TestRegisterOwner_Success(t *testing.T) {
	h := setupHandler(t)

	body := map[string]string{
		"email":        "owner@example.com",
		"password":     "password123",
		"name":         "João Silva",
		"business_name": "Café do João",
		"segment":      "cafeteria",
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.RegisterOwner(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)

	var resp map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))

	// JWT token must NOT be in the JSON body (XSS prevention)
	_, hasToken := resp["token"]
	assert.False(t, hasToken, "token deve estar ausente do JSON body (enviado via cookie HttpOnly)")

	// User data must be present
	user, ok := resp["user"].(map[string]interface{})
	require.True(t, ok, "user field must be present")
	assert.Equal(t, "owner@example.com", user["email"])
	assert.Equal(t, "João Silva", user["name"])
	assert.Equal(t, "owner", user["role"])

	// Business must be present
	biz, ok := resp["business"].(map[string]interface{})
	require.True(t, ok, "business field must be present")
	assert.Equal(t, "Café do João", biz["name"])
	assert.NotEmpty(t, biz["slug"])
}

func TestRegisterOwner_DuplicateEmail(t *testing.T) {
	h := setupHandler(t)

	body := map[string]string{
		"email":        "dup@example.com",
		"password":     "password123",
		"name":         "User One",
		"business_name": "Business One",
		"segment":      "cafeteria",
	}
	b, _ := json.Marshal(body)

	// First registration should succeed
	req1 := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(b))
	req1.Header.Set("Content-Type", "application/json")
	rr1 := httptest.NewRecorder()
	h.RegisterOwner(rr1, req1)
	require.Equal(t, http.StatusCreated, rr1.Code)

	// Second registration with same email should fail with 409
	b2, _ := json.Marshal(body)
	req2 := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(b2))
	req2.Header.Set("Content-Type", "application/json")
	rr2 := httptest.NewRecorder()
	h.RegisterOwner(rr2, req2)

	assert.Equal(t, http.StatusConflict, rr2.Code)

	var errResp map[string]string
	require.NoError(t, json.NewDecoder(rr2.Body).Decode(&errResp))
	assert.Contains(t, errResp["error"], "email")
}

func TestRegisterOwner_MissingFields(t *testing.T) {
	h := setupHandler(t)

	tests := []struct {
		name string
		body map[string]string
	}{
		{
			name: "missing email",
			body: map[string]string{
				"password":     "password123",
				"name":         "Test User",
				"business_name": "Test Business",
				"segment":      "cafeteria",
			},
		},
		{
			name: "missing password",
			body: map[string]string{
				"email":        "test@example.com",
				"name":         "Test User",
				"business_name": "Test Business",
				"segment":      "cafeteria",
			},
		},
		{
			name: "missing name",
			body: map[string]string{
				"email":        "test@example.com",
				"password":     "password123",
				"business_name": "Test Business",
				"segment":      "cafeteria",
			},
		},
		{
			name: "missing business_name",
			body: map[string]string{
				"email":    "test@example.com",
				"password": "password123",
				"name":     "Test User",
				"segment":  "cafeteria",
			},
		},
		{
			name: "password too short",
			body: map[string]string{
				"email":        "test@example.com",
				"password":     "short",
				"name":         "Test User",
				"business_name": "Test Business",
				"segment":      "cafeteria",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b, _ := json.Marshal(tc.body)
			req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(b))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			h.RegisterOwner(rr, req)

			assert.Equal(t, http.StatusBadRequest, rr.Code)
		})
	}
}

// ---- Login tests ----

func TestLogin_Success(t *testing.T) {
	h := setupHandler(t)

	// Seed an owner first
	pool := testutil.SetupTestDB(t)
	q := repository.New(pool)
	testutil.SeedOwner(t, q, "login@example.com", "Login User")

	// But we need to use h's internal DB — so register via the handler
	regBody := map[string]string{
		"email":        "logintest@example.com",
		"password":     "password123",
		"name":         "Login Test",
		"business_name": "Login Business",
		"segment":      "cafeteria",
	}
	b, _ := json.Marshal(regBody)
	regReq := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(b))
	regReq.Header.Set("Content-Type", "application/json")
	regRr := httptest.NewRecorder()
	h.RegisterOwner(regRr, regReq)
	require.Equal(t, http.StatusCreated, regRr.Code)

	// Now login
	loginBody := map[string]string{
		"email":    "logintest@example.com",
		"password": "password123",
	}
	lb, _ := json.Marshal(loginBody)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(lb))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.Login(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))

	// JWT token must NOT be in the JSON body (XSS prevention)
	_, hasToken := resp["token"]
	assert.False(t, hasToken, "token deve estar ausente do JSON body")

	user, ok := resp["user"].(map[string]interface{})
	require.True(t, ok, "user must be present")
	assert.Equal(t, "logintest@example.com", user["email"])
}

func TestLogin_WrongPassword(t *testing.T) {
	h := setupHandler(t)

	// Register first
	regBody := map[string]string{
		"email":        "wrongpw@example.com",
		"password":     "password123",
		"name":         "Wrong PW",
		"business_name": "Wrong PW Business",
		"segment":      "cafeteria",
	}
	b, _ := json.Marshal(regBody)
	regReq := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(b))
	regReq.Header.Set("Content-Type", "application/json")
	regRr := httptest.NewRecorder()
	h.RegisterOwner(regRr, regReq)
	require.Equal(t, http.StatusCreated, regRr.Code)

	// Login with wrong password
	loginBody := map[string]string{
		"email":    "wrongpw@example.com",
		"password": "wrongpassword",
	}
	lb, _ := json.Marshal(loginBody)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(lb))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.Login(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestLogin_UserNotFound(t *testing.T) {
	h := setupHandler(t)

	loginBody := map[string]string{
		"email":    "notfound@example.com",
		"password": "password123",
	}
	lb, _ := json.Marshal(loginBody)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(lb))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.Login(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

// ---- RegisterSubscriber tests ----

func TestRegisterSubscriber_Success(t *testing.T) {
	h := setupHandler(t)

	body := map[string]string{
		"email":    "subscriber@example.com",
		"password": "password123",
		"name":     "Maria Assinante",
		"phone":    "11999998888",
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register-subscriber", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.RegisterSubscriber(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)

	var resp map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))

	// JWT token must NOT be in the JSON body (XSS prevention)
	_, hasToken := resp["token"]
	assert.False(t, hasToken, "token deve estar ausente do JSON body")

	user, ok := resp["user"].(map[string]interface{})
	require.True(t, ok, "user must be present")
	assert.Equal(t, "subscriber@example.com", user["email"])
	assert.Equal(t, "Maria Assinante", user["name"])
	assert.Equal(t, "subscriber", user["role"])
}

func TestRegisterSubscriber_DuplicateEmail(t *testing.T) {
	h := setupHandler(t)

	body := map[string]string{
		"email":    "dupsub@example.com",
		"password": "password123",
		"name":     "Dup Sub",
	}
	b, _ := json.Marshal(body)

	// First registration
	req1 := httptest.NewRequest(http.MethodPost, "/api/auth/register-subscriber", bytes.NewReader(b))
	req1.Header.Set("Content-Type", "application/json")
	rr1 := httptest.NewRecorder()
	h.RegisterSubscriber(rr1, req1)
	require.Equal(t, http.StatusCreated, rr1.Code)

	// Second with same email → 409
	b2, _ := json.Marshal(body)
	req2 := httptest.NewRequest(http.MethodPost, "/api/auth/register-subscriber", bytes.NewReader(b2))
	req2.Header.Set("Content-Type", "application/json")
	rr2 := httptest.NewRecorder()
	h.RegisterSubscriber(rr2, req2)

	assert.Equal(t, http.StatusConflict, rr2.Code)
}

func TestRegisterOwner_ShortPassword(t *testing.T) {
	h := setupHandler(t)

	body := map[string]string{
		"email":         "shortpw@example.com",
		"password":      "short",
		"name":          "Short PW Owner",
		"business_name": "Short PW Café",
		"segment":       "cafeteria",
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.RegisterOwner(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestRegisterOwner_InvalidJSON(t *testing.T) {
	h := setupHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader([]byte(`{invalid json`)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.RegisterOwner(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestLogin_InvalidJSON(t *testing.T) {
	h := setupHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader([]byte(`{invalid json`)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.Login(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestLogin_MissingFields(t *testing.T) {
	h := setupHandler(t)

	tests := []struct {
		name string
		body map[string]string
	}{
		{
			name: "missing email",
			body: map[string]string{
				"password": "password123",
			},
		},
		{
			name: "missing password",
			body: map[string]string{
				"email": "test@example.com",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b, _ := json.Marshal(tc.body)
			req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(b))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			h.Login(rr, req)

			assert.Equal(t, http.StatusBadRequest, rr.Code)
		})
	}
}

func TestRegisterSubscriber_ShortPassword(t *testing.T) {
	h := setupHandler(t)

	body := map[string]string{
		"email":    "subshortpw@example.com",
		"password": "short",
		"name":     "Short PW Sub",
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register-subscriber", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.RegisterSubscriber(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestRegisterSubscriber_InvalidJSON(t *testing.T) {
	h := setupHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register-subscriber", bytes.NewReader([]byte(`{invalid json`)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.RegisterSubscriber(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// ---- HttpOnly Cookie tests ----

func TestLogin_SetsHttpOnlyCookie(t *testing.T) {
	h := setupHandler(t)

	// Register owner first
	regBody := map[string]string{
		"email":         "cookielogin@example.com",
		"password":      "password123",
		"name":          "Cookie Login",
		"business_name": "Cookie Café",
		"segment":       "cafeteria",
	}
	b, _ := json.Marshal(regBody)
	regReq := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(b))
	regReq.Header.Set("Content-Type", "application/json")
	h.RegisterOwner(httptest.NewRecorder(), regReq)

	loginBody := map[string]string{
		"email":    "cookielogin@example.com",
		"password": "password123",
	}
	lb, _ := json.Marshal(loginBody)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(lb))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.Login(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var tokenCookie *http.Cookie
	for _, c := range rr.Result().Cookies() {
		if c.Name == middleware.AuthCookieName {
			tokenCookie = c
			break
		}
	}
	require.NotNil(t, tokenCookie, "deve definir cookie clubepay_token")
	assert.True(t, tokenCookie.HttpOnly, "cookie deve ser HttpOnly")
	assert.True(t, tokenCookie.Secure, "cookie deve ter flag Secure")
	assert.NotEmpty(t, tokenCookie.Value)
	assert.Equal(t, "/", tokenCookie.Path)
	assert.Greater(t, tokenCookie.MaxAge, 0)
}

func TestRegisterOwner_SetsHttpOnlyCookie(t *testing.T) {
	h := setupHandler(t)

	body := map[string]string{
		"email":         "cookiereg@example.com",
		"password":      "password123",
		"name":          "Cookie Reg",
		"business_name": "Cookie Reg Café",
		"segment":       "cafeteria",
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.RegisterOwner(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)

	var tokenCookie *http.Cookie
	for _, c := range rr.Result().Cookies() {
		if c.Name == middleware.AuthCookieName {
			tokenCookie = c
			break
		}
	}
	require.NotNil(t, tokenCookie, "deve definir cookie clubepay_token")
	assert.True(t, tokenCookie.HttpOnly, "cookie deve ser HttpOnly")
	assert.True(t, tokenCookie.Secure, "cookie deve ter flag Secure")
	assert.NotEmpty(t, tokenCookie.Value)
}

func TestRegisterSubscriber_SetsHttpOnlyCookie(t *testing.T) {
	h := setupHandler(t)

	body := map[string]string{
		"email":    "cookiesub@example.com",
		"password": "password123",
		"name":     "Cookie Sub",
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register-subscriber", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.RegisterSubscriber(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)

	var tokenCookie *http.Cookie
	for _, c := range rr.Result().Cookies() {
		if c.Name == middleware.AuthCookieName {
			tokenCookie = c
			break
		}
	}
	require.NotNil(t, tokenCookie, "deve definir cookie clubepay_token")
	assert.True(t, tokenCookie.HttpOnly, "cookie deve ser HttpOnly")
	assert.True(t, tokenCookie.Secure, "cookie deve ter flag Secure")
	assert.NotEmpty(t, tokenCookie.Value)
}

func TestRegisterSubscriber_MissingFields(t *testing.T) {
	h := setupHandler(t)

	tests := []struct {
		name string
		body map[string]string
	}{
		{
			name: "missing email",
			body: map[string]string{
				"password": "password123",
				"name":     "Test Sub",
			},
		},
		{
			name: "missing password",
			body: map[string]string{
				"email": "test@example.com",
				"name":  "Test Sub",
			},
		},
		{
			name: "missing name",
			body: map[string]string{
				"email":    "test@example.com",
				"password": "password123",
			},
		},
		{
			name: "password too short",
			body: map[string]string{
				"email":    "test@example.com",
				"password": "short",
				"name":     "Test Sub",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b, _ := json.Marshal(tc.body)
			req := httptest.NewRequest(http.MethodPost, "/api/auth/register-subscriber", bytes.NewReader(b))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			h.RegisterSubscriber(rr, req)

			assert.Equal(t, http.StatusBadRequest, rr.Code)
		})
	}
}
