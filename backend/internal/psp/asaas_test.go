package psp_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/clubepay/backend/internal/psp"
)

func TestAsaas_CreateCustomer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/customers", r.URL.Path)
		assert.Equal(t, "test-api-key", r.Header.Get("access_token"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Read and verify body
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		defer r.Body.Close()

		var reqBody map[string]string
		require.NoError(t, json.Unmarshal(body, &reqBody))
		assert.Equal(t, "Maria Silva", reqBody["name"])
		assert.Equal(t, "maria@test.com", reqBody["email"])
		assert.Equal(t, "12345678901", reqBody["cpfCnpj"])

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"id": "cus_abc123"})
	}))
	defer server.Close()

	client := psp.NewAsaas(server.URL, "test-api-key", "")

	customer, err := client.CreateCustomer(context.Background(), psp.CreateCustomerRequest{
		Name:  "Maria Silva",
		Email: "maria@test.com",
		CPF:   "12345678901",
	})

	require.NoError(t, err)
	assert.Equal(t, "cus_abc123", customer.ID)
}

func TestAsaas_CreateSubscription(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/subscriptions", r.URL.Path)
		assert.Equal(t, "test-api-key", r.Header.Get("access_token"))

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		defer r.Body.Close()

		var reqBody map[string]interface{}
		require.NoError(t, json.Unmarshal(body, &reqBody))
		assert.Equal(t, "cus_abc123", reqBody["customer"])
		assert.Equal(t, "PIX", reqBody["billingType"])
		assert.Equal(t, 29.90, reqBody["value"]) // 2990 cents → 29.90 reais
		assert.Equal(t, "MONTHLY", reqBody["cycle"])
		assert.Equal(t, "Plano Cafe", reqBody["description"])

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"id":     "sub_xyz789",
			"status": "ACTIVE",
		})
	}))
	defer server.Close()

	client := psp.NewAsaas(server.URL, "test-api-key", "")

	sub, err := client.CreateSubscription(context.Background(), psp.CreateSubscriptionRequest{
		CustomerID:  "cus_abc123",
		PriceCents:  2990,
		Description: "Plano Cafe",
		Cycle:       "MONTHLY",
	})

	require.NoError(t, err)
	assert.Equal(t, "sub_xyz789", sub.ID)
	assert.Equal(t, "ACTIVE", sub.Status)
}

func TestAsaas_CancelSubscription(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/subscriptions/sub_cancel_123", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"deleted": true})
	}))
	defer server.Close()

	client := psp.NewAsaas(server.URL, "test-api-key", "")

	err := client.CancelSubscription(context.Background(), "sub_cancel_123")
	require.NoError(t, err)
}

func TestAsaas_GetSubscription(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/subscriptions/sub_get_123", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"id":     "sub_get_123",
			"status": "ACTIVE",
		})
	}))
	defer server.Close()

	client := psp.NewAsaas(server.URL, "test-api-key", "")

	sub, err := client.GetSubscription(context.Background(), "sub_get_123")
	require.NoError(t, err)
	assert.Equal(t, "sub_get_123", sub.ID)
	assert.Equal(t, "ACTIVE", sub.Status)
}

func TestAsaas_GetPayments(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/subscriptions/sub_pay_123/payments", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]string{
				{"id": "pay_1", "status": "CONFIRMED"},
				{"id": "pay_2", "status": "PENDING"},
			},
		})
	}))
	defer server.Close()

	client := psp.NewAsaas(server.URL, "test-api-key", "")

	payments, err := client.GetPayments(context.Background(), "sub_pay_123")
	require.NoError(t, err)
	require.Len(t, payments, 2)
	assert.Equal(t, "pay_1", payments[0].ID)
	assert.Equal(t, "CONFIRMED", payments[0].Status)
	assert.Equal(t, "pay_2", payments[1].ID)
	assert.Equal(t, "PENDING", payments[1].Status)
}

func TestAsaas_ValidateWebhook(t *testing.T) {
	secret := "my-webhook-secret"
	client := psp.NewAsaas("http://localhost", "key", secret)

	payload := []byte(`{"event":"PAYMENT_CONFIRMED"}`)

	// Compute valid HMAC
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	validSig := hex.EncodeToString(mac.Sum(nil))

	assert.True(t, client.ValidateWebhook(payload, validSig))
	assert.False(t, client.ValidateWebhook(payload, "invalid-signature"))
}

func TestAsaas_ValidateWebhook_NoSecret(t *testing.T) {
	client := psp.NewAsaas("http://localhost", "key", "")

	// With no secret configured, always returns true
	assert.True(t, client.ValidateWebhook([]byte("anything"), ""))
}

func TestAsaas_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"errors":[{"code":"invalid_value"}]}`))
	}))
	defer server.Close()

	client := psp.NewAsaas(server.URL, "test-api-key", "")

	_, err := client.CreateCustomer(context.Background(), psp.CreateCustomerRequest{
		Name:  "Test",
		Email: "test@test.com",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "asaas API error")
}
