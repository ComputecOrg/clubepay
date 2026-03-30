package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/clubepay/backend/internal/middleware"
	"github.com/stretchr/testify/assert"
)

func TestRateLimit_AllowsUnderLimit(t *testing.T) {
	handler := middleware.RateLimit(5, time.Minute)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("POST", "/api/auth/login", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code, "request %d should be allowed", i+1)
	}
}

func TestRateLimit_BlocksOverLimit(t *testing.T) {
	handler := middleware.RateLimit(2, time.Minute)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	ip := "10.0.0.1:9999"

	// First two requests should succeed
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("POST", "/api/auth/login", nil)
		req.RemoteAddr = ip
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code, "request %d should be allowed", i+1)
	}

	// Third request should be blocked
	req := httptest.NewRequest("POST", "/api/auth/login", nil)
	req.RemoteAddr = ip
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusTooManyRequests, rr.Code, "3rd request should be blocked")
	assert.NotEmpty(t, rr.Header().Get("Retry-After"), "Retry-After header should be set on 429")
}

func TestRateLimit_DifferentIPs(t *testing.T) {
	handler := middleware.RateLimit(2, time.Minute)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Exhaust limit for IP A
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("POST", "/api/auth/login", nil)
		req.RemoteAddr = "1.1.1.1:1000"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	}

	// IP A should now be blocked
	reqA := httptest.NewRequest("POST", "/api/auth/login", nil)
	reqA.RemoteAddr = "1.1.1.1:1000"
	rrA := httptest.NewRecorder()
	handler.ServeHTTP(rrA, reqA)
	assert.Equal(t, http.StatusTooManyRequests, rrA.Code, "IP A should be blocked")

	// IP B should still be allowed (separate counter)
	reqB := httptest.NewRequest("POST", "/api/auth/login", nil)
	reqB.RemoteAddr = "2.2.2.2:2000"
	rrB := httptest.NewRecorder()
	handler.ServeHTTP(rrB, reqB)
	assert.Equal(t, http.StatusOK, rrB.Code, "IP B should be allowed independently")
}

func TestRateLimit_WindowReset(t *testing.T) {
	// Use a very short window so we can test expiry
	handler := middleware.RateLimit(1, 50*time.Millisecond)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	ip := "3.3.3.3:3000"

	// First request OK
	req := httptest.NewRequest("POST", "/", nil)
	req.RemoteAddr = ip
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	// Second request blocked
	req = httptest.NewRequest("POST", "/", nil)
	req.RemoteAddr = ip
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusTooManyRequests, rr.Code)

	// Wait for window to expire
	time.Sleep(100 * time.Millisecond)

	// Should be allowed again after window reset
	req = httptest.NewRequest("POST", "/", nil)
	req.RemoteAddr = ip
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code, "should be allowed after window expired")
}
