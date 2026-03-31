package analytics

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

const (
	ga4CollectURL = "https://www.google-analytics.com/mp/collect"
)

// GA4Client implements Client interface for Google Analytics 4
type GA4Client struct {
	measurementID string
	apiSecret     string
	logger        *slog.Logger
	httpClient    *http.Client
}

// NewGA4Client creates a new GA4 client
func NewGA4Client(measurementID, apiSecret string) (*GA4Client, error) {
	if measurementID == "" || apiSecret == "" {
		return nil, fmt.Errorf("measurement_id and api_secret are required")
	}

	return &GA4Client{
		measurementID: measurementID,
		apiSecret:     apiSecret,
		logger:        slog.Default(),
		httpClient:    &http.Client{},
	}, nil
}

// TrackBusinessCreated tracks when a business owner creates their business
func (c *GA4Client) TrackBusinessCreated(businessID, email, industry, city string) error {
	return c.sendEvent("business_created", anonymizeEmail(email), map[string]interface{}{
		"business_id": businessID,
		"industry":    industry,
		"city":        city,
	})
}

// TrackSubscriptionCreated tracks when a subscriber subscribes to a plan
func (c *GA4Client) TrackSubscriptionCreated(subID, planID, businessID string, amount int) error {
	return c.sendEvent("subscription_created", subID, map[string]interface{}{
		"plan_id":      planID,
		"business_id":  businessID,
		"amount":       amount,
		"currency":     "BRL",
	})
}

// TrackValidationSuccess tracks when a subscriber validates usage successfully
func (c *GA4Client) TrackValidationSuccess(subID, planID string, count int) error {
	return c.sendEvent("validation_success", subID, map[string]interface{}{
		"plan_id": planID,
		"count":   count,
	})
}

// TrackValidationFailed tracks when a subscriber's validation fails
func (c *GA4Client) TrackValidationFailed(subID, reason string) error {
	return c.sendEvent("validation_failed", subID, map[string]interface{}{
		"reason": reason,
	})
}

// TrackPaymentReceived tracks when a payment is successfully received
func (c *GA4Client) TrackPaymentReceived(subID string, amount int, planID string) error {
	return c.sendEvent("payment_received", subID, map[string]interface{}{
		"amount":   amount,
		"plan_id":  planID,
		"currency": "BRL",
	})
}

// TrackPaymentFailed tracks when a payment fails
func (c *GA4Client) TrackPaymentFailed(subID, reason string) error {
	return c.sendEvent("payment_failed", subID, map[string]interface{}{
		"reason": reason,
	})
}

// TrackSubscriptionCanceled tracks when a subscription is canceled
func (c *GA4Client) TrackSubscriptionCanceled(subID, reason string) error {
	return c.sendEvent("subscription_canceled", subID, map[string]interface{}{
		"reason": reason,
	})
}

// sendEvent sends an event to GA4 (internal helper)
func (c *GA4Client) sendEvent(eventName, clientID string, params map[string]interface{}) error {
	payload := map[string]interface{}{
		"client_id": clientID,
		"events": []map[string]interface{}{
			{
				"name":   eventName,
				"params": params,
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		c.logger.Error("failed to marshal GA4 payload", "event", eventName, "error", err)
		return err
	}

	url := fmt.Sprintf("%s?measurement_id=%s&api_secret=%s", ga4CollectURL, c.measurementID, c.apiSecret)

	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		c.logger.Error("failed to create GA4 request", "event", eventName, "error", err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("failed to send GA4 request", "event", eventName, "error", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		c.logger.Error("GA4 returned error", "status", resp.StatusCode, "body", string(respBody), "event", eventName)
		return fmt.Errorf("GA4 returned status %d", resp.StatusCode)
	}

	c.logger.Debug("GA4 event tracked", "event", eventName, "client_id", clientID)
	return nil
}

// anonymizeEmail converts email to SHA256 hash (GDPR-compliant)
func anonymizeEmail(email string) string {
	hash := sha256.Sum256([]byte(email))
	return fmt.Sprintf("%x", hash)
}
