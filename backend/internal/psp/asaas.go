package psp

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Asaas implements the PSP interface for the Asaas payment provider.
type Asaas struct {
	baseURL       string
	apiKey        string
	webhookSecret string
	client        *http.Client
}

// NewAsaas creates a new Asaas PSP client.
func NewAsaas(baseURL, apiKey, webhookSecret string) *Asaas {
	return &Asaas{
		baseURL:       baseURL,
		apiKey:        apiKey,
		webhookSecret: webhookSecret,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CreateCustomer creates a customer in Asaas.
// POST /customers
func (a *Asaas) CreateCustomer(ctx context.Context, req CreateCustomerRequest) (*Customer, error) {
	body := map[string]string{
		"name":    req.Name,
		"email":   req.Email,
		"cpfCnpj": req.CPF,
	}
	if req.Phone != "" {
		body["phone"] = req.Phone
	}

	respBody, err := a.doRequest(ctx, http.MethodPost, "/customers", body)
	if err != nil {
		return nil, fmt.Errorf("create customer: %w", err)
	}

	var result struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parse customer response: %w", err)
	}

	return &Customer{ID: result.ID}, nil
}

// CreateSubscription creates a subscription in Asaas.
// POST /subscriptions
func (a *Asaas) CreateSubscription(ctx context.Context, req CreateSubscriptionRequest) (*Subscription, error) {
	// Convert cents to reais (Asaas expects float value in BRL)
	valueReais := float64(req.PriceCents) / 100.0

	body := map[string]interface{}{
		"customer":    req.CustomerID,
		"billingType": "PIX",
		"value":       valueReais,
		"cycle":       req.Cycle,
		"description": req.Description,
	}

	respBody, err := a.doRequest(ctx, http.MethodPost, "/subscriptions", body)
	if err != nil {
		return nil, fmt.Errorf("create subscription: %w", err)
	}

	var result struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parse subscription response: %w", err)
	}

	return &Subscription{ID: result.ID, Status: result.Status}, nil
}

// CancelSubscription cancels a subscription in Asaas.
// DELETE /subscriptions/{id}
func (a *Asaas) CancelSubscription(ctx context.Context, subscriptionID string) error {
	_, err := a.doRequest(ctx, http.MethodDelete, "/subscriptions/"+subscriptionID, nil)
	if err != nil {
		return fmt.Errorf("cancel subscription: %w", err)
	}
	return nil
}

// GetSubscription fetches subscription details from Asaas.
// GET /subscriptions/{id}
func (a *Asaas) GetSubscription(ctx context.Context, subscriptionID string) (*Subscription, error) {
	respBody, err := a.doRequest(ctx, http.MethodGet, "/subscriptions/"+subscriptionID, nil)
	if err != nil {
		return nil, fmt.Errorf("get subscription: %w", err)
	}

	var result struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parse subscription response: %w", err)
	}

	return &Subscription{ID: result.ID, Status: result.Status}, nil
}

// GetPayments fetches payments for a subscription from Asaas.
// GET /subscriptions/{id}/payments
func (a *Asaas) GetPayments(ctx context.Context, subscriptionID string) ([]Payment, error) {
	respBody, err := a.doRequest(ctx, http.MethodGet, "/subscriptions/"+subscriptionID+"/payments", nil)
	if err != nil {
		return nil, fmt.Errorf("get payments: %w", err)
	}

	var result struct {
		Data []struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parse payments response: %w", err)
	}

	payments := make([]Payment, len(result.Data))
	for i, p := range result.Data {
		payments[i] = Payment{ID: p.ID, Status: p.Status}
	}

	return payments, nil
}

// ValidateWebhook validates the HMAC SHA256 signature of a webhook payload.
func (a *Asaas) ValidateWebhook(payload []byte, signature string) bool {
	if a.webhookSecret == "" {
		return true
	}
	mac := hmac.New(sha256.New, []byte(a.webhookSecret))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

// doRequest performs an authenticated HTTP request to the Asaas API.
func (a *Asaas) doRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, a.baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("access_token", a.apiKey)

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("asaas API error: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}
