package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func dummyHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestCORS_AllowedOrigin(t *testing.T) {
	allowedOrigin := "https://app.clubepay.com"
	handler := CORS(allowedOrigin)(http.HandlerFunc(dummyHandler))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", allowedOrigin)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != allowedOrigin {
		t.Errorf("expected Access-Control-Allow-Origin %q, got %q", allowedOrigin, got)
	}
	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}

func TestCORS_DisallowedOrigin(t *testing.T) {
	allowedOrigin := "https://app.clubepay.com"
	handler := CORS(allowedOrigin)(http.HandlerFunc(dummyHandler))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://evil.example.com")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("expected no Access-Control-Allow-Origin header, got %q", got)
	}
}

func TestCORS_Preflight(t *testing.T) {
	allowedOrigin := "https://app.clubepay.com"
	handler := CORS(allowedOrigin)(http.HandlerFunc(dummyHandler))

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", allowedOrigin)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200 for preflight, got %d", rr.Code)
	}
	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != allowedOrigin {
		t.Errorf("expected Access-Control-Allow-Origin %q, got %q", allowedOrigin, got)
	}
	if got := rr.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Error("expected Access-Control-Allow-Methods to be set")
	}
	if got := rr.Header().Get("Access-Control-Allow-Headers"); got == "" {
		t.Error("expected Access-Control-Allow-Headers to be set")
	}
}

func TestCORS_WildcardForDev(t *testing.T) {
	handler := CORS("*")(http.HandlerFunc(dummyHandler))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://anything.example.com")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("expected Access-Control-Allow-Origin \"*\", got %q", got)
	}
	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}

func TestCORS_MultipleAllowedOrigins(t *testing.T) {
	allowedOrigins := "https://app.clubepay.com,https://staging.clubepay.com"
	handler := CORS(allowedOrigins)(http.HandlerFunc(dummyHandler))

	for _, origin := range []string{"https://app.clubepay.com", "https://staging.clubepay.com"} {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Origin", origin)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if got := rr.Header().Get("Access-Control-Allow-Origin"); got != origin {
			t.Errorf("origin %q: expected Access-Control-Allow-Origin %q, got %q", origin, origin, got)
		}
	}
}
