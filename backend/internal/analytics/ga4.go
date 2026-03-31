package analytics

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// GA4EventTracker implementa EventTracker usando Google Analytics 4
type GA4EventTracker struct {
	measurementID string
	apiSecret     string
	userID        string
	client        *http.Client
	eventBuffer   []Event
	bufferMutex   sync.Mutex
	flushTicker   *time.Ticker
	stopChan      chan struct{}
}

// GA4Payload é a estrutura enviada para GA4
type GA4Payload struct {
	ClientID string          `json:"client_id"`
	UserID   string          `json:"user_id,omitempty"`
	Events   []GA4EventData  `json:"events"`
}

// GA4EventData é o formato de evento esperado pelo GA4
type GA4EventData struct {
	Name       string                 `json:"name"`
	Params     map[string]interface{} `json:"params"`
	Timestamp  string                 `json:"timestamp,omitempty"`
}

const (
	ga4APIEndpoint   = "https://www.google-analytics.com/mp/collect"
	defaultFlushTime = 30 * time.Second
)

// NewGA4EventTracker cria um novo GA4EventTracker
func NewGA4EventTracker(measurementID, apiSecret, userID string) *GA4EventTracker {
	tracker := &GA4EventTracker{
		measurementID: measurementID,
		apiSecret:     apiSecret,
		userID:        userID,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		eventBuffer: make([]Event, 0),
		stopChan:    make(chan struct{}),
	}

	// Inicia ticker para flush automático (apenas se config valida)
	if measurementID != "" && apiSecret != "" {
		tracker.flushTicker = time.NewTicker(defaultFlushTime)
		go tracker.autoFlush()
	}

	return tracker
}

// TrackBusinessCreated implementa EventTracker.TrackBusinessCreated
func (g *GA4EventTracker) TrackBusinessCreated(ownerID, businessName, industry, city string) error {
	event := Event{
		Name: "business_created",
		Params: map[string]interface{}{
			"user_id":       ownerID,
			"business_name": businessName,
			"industry":      industry,
			"city":          city,
		},
	}
	return g.trackEvent(event)
}

// TrackFirstPlanCreated implementa EventTracker.TrackFirstPlanCreated
func (g *GA4EventTracker) TrackFirstPlanCreated(ownerID, planID, planName string, priceInCents int) error {
	event := Event{
		Name: "first_plan_created",
		Params: map[string]interface{}{
			"user_id":   ownerID,
			"plan_id":   planID,
			"plan_name": planName,
			"price":     float64(priceInCents) / 100,
		},
	}
	return g.trackEvent(event)
}

// TrackFirstSubscriber implementa EventTracker.TrackFirstSubscriber
func (g *GA4EventTracker) TrackFirstSubscriber(ownerID, planID string) error {
	event := Event{
		Name: "first_subscriber",
		Params: map[string]interface{}{
			"user_id": ownerID,
			"plan_id": planID,
		},
	}
	return g.trackEvent(event)
}

// TrackSubscriptionCanceled implementa EventTracker.TrackSubscriptionCanceled
func (g *GA4EventTracker) TrackSubscriptionCanceled(ownerID, subscriptionID, reason string) error {
	event := Event{
		Name: "subscription_canceled",
		Params: map[string]interface{}{
			"user_id":         ownerID,
			"subscription_id": subscriptionID,
			"reason":          reason,
		},
	}
	return g.trackEvent(event)
}

// TrackSubscriptionCreated implementa EventTracker.TrackSubscriptionCreated
func (g *GA4EventTracker) TrackSubscriptionCreated(subscriberID, planID, planName, referralCode string) error {
	params := map[string]interface{}{
		"user_id":   subscriberID,
		"plan_id":   planID,
		"plan_name": planName,
	}
	if referralCode != "" {
		params["referral_code"] = referralCode
	}

	event := Event{
		Name:   "subscription_created",
		Params: params,
	}
	return g.trackEvent(event)
}

// TrackValidationSuccess implementa EventTracker.TrackValidationSuccess
func (g *GA4EventTracker) TrackValidationSuccess(subscriberID, planID string, dailyCount int) error {
	event := Event{
		Name: "validation_success",
		Params: map[string]interface{}{
			"user_id":    subscriberID,
			"plan_id":    planID,
			"daily_count": float64(dailyCount),
		},
	}
	return g.trackEvent(event)
}

// TrackValidationFailed implementa EventTracker.TrackValidationFailed
func (g *GA4EventTracker) TrackValidationFailed(subscriberID, reason string) error {
	event := Event{
		Name: "validation_failed",
		Params: map[string]interface{}{
			"user_id": subscriberID,
			"reason":  reason,
		},
	}
	return g.trackEvent(event)
}

// TrackPlanUpgraded implementa EventTracker.TrackPlanUpgraded
func (g *GA4EventTracker) TrackPlanUpgraded(subscriberID, fromPlanID, toPlanID string) error {
	event := Event{
		Name: "plan_upgraded",
		Params: map[string]interface{}{
			"user_id":      subscriberID,
			"from_plan_id": fromPlanID,
			"to_plan_id":   toPlanID,
		},
	}
	return g.trackEvent(event)
}

// TrackCancellation implementa EventTracker.TrackCancellation
func (g *GA4EventTracker) TrackCancellation(subscriberID, reason string) error {
	event := Event{
		Name: "cancellation",
		Params: map[string]interface{}{
			"user_id": subscriberID,
			"reason":  reason,
		},
	}
	return g.trackEvent(event)
}

// TrackPaymentReceived implementa EventTracker.TrackPaymentReceived
func (g *GA4EventTracker) TrackPaymentReceived(ownerID, subscriptionID string, amountInCents int, planType string) error {
	event := Event{
		Name: "payment_received",
		Params: map[string]interface{}{
			"user_id":         ownerID,
			"subscription_id": subscriptionID,
			"amount":          float64(amountInCents) / 100,
			"plan_type":       planType,
		},
	}
	return g.trackEvent(event)
}

// TrackPaymentFailed implementa EventTracker.TrackPaymentFailed
func (g *GA4EventTracker) TrackPaymentFailed(ownerID, subscriptionID, reason string) error {
	event := Event{
		Name: "payment_failed",
		Params: map[string]interface{}{
			"user_id":         ownerID,
			"subscription_id": subscriptionID,
			"reason":          reason,
		},
	}
	return g.trackEvent(event)
}

// trackEvent adiciona evento ao buffer
func (g *GA4EventTracker) trackEvent(event Event) error {
	g.bufferMutex.Lock()
	defer g.bufferMutex.Unlock()

	g.eventBuffer = append(g.eventBuffer, event)

	// Flush automático se buffer atingir tamanho crítico
	if len(g.eventBuffer) >= 50 {
		return g.flushLocked()
	}

	return nil
}

// Flush envia eventos pendentes para GA4
func (g *GA4EventTracker) Flush(ctx context.Context) error {
	g.bufferMutex.Lock()
	defer g.bufferMutex.Unlock()

	return g.flushLocked()
}

// flushLocked envia eventos sem locks (deve ser chamado com mutex travado)
func (g *GA4EventTracker) flushLocked() error {
	if len(g.eventBuffer) == 0 {
		return nil
	}

	// Se measurement ID não está configurado, apenas descarta eventos
	if g.measurementID == "" || g.apiSecret == "" {
		g.eventBuffer = make([]Event, 0)
		return nil
	}

	// Converte eventos para formato GA4
	ga4Events := make([]GA4EventData, len(g.eventBuffer))
	for i, event := range g.eventBuffer {
		ga4Events[i] = GA4EventData{
			Name:      event.Name,
			Params:    event.Params,
			Timestamp: fmt.Sprintf("%d", time.Now().UnixMilli()),
		}
	}

	// Cria payload
	payload := GA4Payload{
		ClientID: g.userID, // Usando userID como clientID para simplicidade
		UserID:   g.userID,
		Events:   ga4Events,
	}

	// Serializa para JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal GA4 payload: %w", err)
	}

	// Envia para GA4
	url := fmt.Sprintf("%s?measurement_id=%s&api_secret=%s", ga4APIEndpoint, g.measurementID, g.apiSecret)
	resp, err := g.client.Post(url, "application/json", bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send events to GA4: %w", err)
	}
	defer resp.Body.Close()

	// Log resposta se houver erro
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GA4 API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Limpa buffer
	g.eventBuffer = make([]Event, 0)

	return nil
}

// autoFlush envia eventos periodicamente
func (g *GA4EventTracker) autoFlush() {
	if g.flushTicker == nil {
		return
	}

	for {
		select {
		case <-g.flushTicker.C:
			_ = g.Flush(context.Background())
		case <-g.stopChan:
			g.flushTicker.Stop()
			return
		}
	}
}

// Close encerra o tracker e faz flush final
func (g *GA4EventTracker) Close(ctx context.Context) error {
	if g.flushTicker != nil {
		close(g.stopChan)
	}
	return g.Flush(ctx)
}
