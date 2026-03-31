package analytics

import (
	"context"
)

// Event representa um evento de analytics para ser rastreado
type Event struct {
	Name   string                 `json:"name"`
	Params map[string]interface{} `json:"params"`
}

// EventTracker interface define os métodos para rastrear eventos
type EventTracker interface {
	// Business events
	TrackBusinessCreated(ownerID, businessName, industry, city string) error
	TrackFirstPlanCreated(ownerID, planID, planName string, priceInCents int) error
	TrackFirstSubscriber(ownerID, planID string) error
	TrackSubscriptionCanceled(ownerID, subscriptionID, reason string) error

	// Subscriber events
	TrackSubscriptionCreated(subscriberID, planID, planName, referralCode string) error
	TrackValidationSuccess(subscriberID, planID string, dailyCount int) error
	TrackValidationFailed(subscriberID, reason string) error
	TrackPlanUpgraded(subscriberID, fromPlanID, toPlanID string) error
	TrackCancellation(subscriberID, reason string) error

	// Revenue events
	TrackPaymentReceived(ownerID, subscriptionID string, amountInCents int, planType string) error
	TrackPaymentFailed(ownerID, subscriptionID, reason string) error

	// Flush any pending events
	Flush(ctx context.Context) error
}

// MockEventTracker é uma implementação mock para testes
type MockEventTracker struct {
	recordedEvents []Event
}

// NewMockEventTracker cria um novo MockEventTracker
func NewMockEventTracker() *MockEventTracker {
	return &MockEventTracker{
		recordedEvents: []Event{},
	}
}

// TrackBusinessCreated registra evento de negócio criado
func (m *MockEventTracker) TrackBusinessCreated(ownerID, businessName, industry, city string) error {
	m.recordedEvents = append(m.recordedEvents, Event{
		Name: "business_created",
		Params: map[string]interface{}{
			"user_id":      ownerID,
			"business_name": businessName,
			"industry":     industry,
			"city":         city,
		},
	})
	return nil
}

// TrackFirstPlanCreated registra evento do primeiro plano criado
func (m *MockEventTracker) TrackFirstPlanCreated(ownerID, planID, planName string, priceInCents int) error {
	m.recordedEvents = append(m.recordedEvents, Event{
		Name: "first_plan_created",
		Params: map[string]interface{}{
			"user_id":    ownerID,
			"plan_id":    planID,
			"plan_name":  planName,
			"price":      float64(priceInCents) / 100,
		},
	})
	return nil
}

// TrackFirstSubscriber registra evento do primeiro assinante
func (m *MockEventTracker) TrackFirstSubscriber(ownerID, planID string) error {
	m.recordedEvents = append(m.recordedEvents, Event{
		Name: "first_subscriber",
		Params: map[string]interface{}{
			"user_id": ownerID,
			"plan_id": planID,
		},
	})
	return nil
}

// TrackSubscriptionCanceled registra cancelamento de assinatura
func (m *MockEventTracker) TrackSubscriptionCanceled(ownerID, subscriptionID, reason string) error {
	m.recordedEvents = append(m.recordedEvents, Event{
		Name: "subscription_canceled",
		Params: map[string]interface{}{
			"user_id":         ownerID,
			"subscription_id": subscriptionID,
			"reason":          reason,
		},
	})
	return nil
}

// TrackSubscriptionCreated registra evento de assinatura criada
func (m *MockEventTracker) TrackSubscriptionCreated(subscriberID, planID, planName, referralCode string) error {
	params := map[string]interface{}{
		"user_id":   subscriberID,
		"plan_id":   planID,
		"plan_name": planName,
	}
	if referralCode != "" {
		params["referral_code"] = referralCode
	}

	m.recordedEvents = append(m.recordedEvents, Event{
		Name:   "subscription_created",
		Params: params,
	})
	return nil
}

// TrackValidationSuccess registra sucesso na validação
func (m *MockEventTracker) TrackValidationSuccess(subscriberID, planID string, dailyCount int) error {
	m.recordedEvents = append(m.recordedEvents, Event{
		Name: "validation_success",
		Params: map[string]interface{}{
			"user_id":    subscriberID,
			"plan_id":    planID,
			"daily_count": float64(dailyCount),
		},
	})
	return nil
}

// TrackValidationFailed registra falha na validação
func (m *MockEventTracker) TrackValidationFailed(subscriberID, reason string) error {
	m.recordedEvents = append(m.recordedEvents, Event{
		Name: "validation_failed",
		Params: map[string]interface{}{
			"user_id": subscriberID,
			"reason":  reason,
		},
	})
	return nil
}

// TrackPlanUpgraded registra upgrade de plano
func (m *MockEventTracker) TrackPlanUpgraded(subscriberID, fromPlanID, toPlanID string) error {
	m.recordedEvents = append(m.recordedEvents, Event{
		Name: "plan_upgraded",
		Params: map[string]interface{}{
			"user_id":      subscriberID,
			"from_plan_id": fromPlanID,
			"to_plan_id":   toPlanID,
		},
	})
	return nil
}

// TrackCancellation registra cancelamento
func (m *MockEventTracker) TrackCancellation(subscriberID, reason string) error {
	m.recordedEvents = append(m.recordedEvents, Event{
		Name: "cancellation",
		Params: map[string]interface{}{
			"user_id": subscriberID,
			"reason":  reason,
		},
	})
	return nil
}

// TrackPaymentReceived registra pagamento recebido
func (m *MockEventTracker) TrackPaymentReceived(ownerID, subscriptionID string, amountInCents int, planType string) error {
	m.recordedEvents = append(m.recordedEvents, Event{
		Name: "payment_received",
		Params: map[string]interface{}{
			"user_id":         ownerID,
			"subscription_id": subscriptionID,
			"amount":          float64(amountInCents) / 100,
			"plan_type":       planType,
		},
	})
	return nil
}

// TrackPaymentFailed registra falha em pagamento
func (m *MockEventTracker) TrackPaymentFailed(ownerID, subscriptionID, reason string) error {
	m.recordedEvents = append(m.recordedEvents, Event{
		Name: "payment_failed",
		Params: map[string]interface{}{
			"user_id":         ownerID,
			"subscription_id": subscriptionID,
			"reason":          reason,
		},
	})
	return nil
}

// Flush aguarda e envia eventos pendentes
func (m *MockEventTracker) Flush(ctx context.Context) error {
	return nil
}

// GetRecordedEvents retorna eventos registrados (apenas para testes)
func (m *MockEventTracker) GetRecordedEvents() []Event {
	return m.recordedEvents
}
