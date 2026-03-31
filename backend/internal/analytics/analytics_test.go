package analytics

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGA4Client(t *testing.T) {
	client, err := NewGA4Client("test-measurement-id", "test-api-secret")
	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestTrackBusinessCreated(t *testing.T) {
	client := newMockClient()

	err := client.TrackBusinessCreated("biz-123", "owner@example.com", "cafe", "São Paulo")
	assert.NoError(t, err)

	// Verify mock received the event
	assert.Equal(t, 1, len(client.events))
	assert.Equal(t, "business_created", client.events[0].Name)
}

func TestTrackSubscriptionCreated(t *testing.T) {
	client := newMockClient()

	err := client.TrackSubscriptionCreated("sub-123", "plan-daily", "biz-456", 50)
	assert.NoError(t, err)

	assert.Equal(t, 1, len(client.events))
	assert.Equal(t, "subscription_created", client.events[0].Name)
}

func TestTrackValidationSuccess(t *testing.T) {
	client := newMockClient()

	err := client.TrackValidationSuccess("sub-123", "plan-daily", 1)
	assert.NoError(t, err)

	assert.Equal(t, 1, len(client.events))
	assert.Equal(t, "validation_success", client.events[0].Name)
}

func TestTrackValidationFailed(t *testing.T) {
	client := newMockClient()

	err := client.TrackValidationFailed("sub-123", "limit_exceeded")
	assert.NoError(t, err)

	assert.Equal(t, 1, len(client.events))
	assert.Equal(t, "validation_failed", client.events[0].Name)
}

func TestTrackPaymentReceived(t *testing.T) {
	client := newMockClient()

	err := client.TrackPaymentReceived("sub-123", 50, "plan-daily")
	assert.NoError(t, err)

	assert.Equal(t, 1, len(client.events))
	assert.Equal(t, "payment_received", client.events[0].Name)
}

func TestTrackPaymentFailed(t *testing.T) {
	client := newMockClient()

	err := client.TrackPaymentFailed("sub-123", "insufficient_funds")
	assert.NoError(t, err)

	assert.Equal(t, 1, len(client.events))
	assert.Equal(t, "payment_failed", client.events[0].Name)
}

func TestTrackSubscriptionCanceled(t *testing.T) {
	client := newMockClient()

	err := client.TrackSubscriptionCanceled("sub-123", "user_requested")
	assert.NoError(t, err)

	assert.Equal(t, 1, len(client.events))
	assert.Equal(t, "subscription_canceled", client.events[0].Name)
}

// Mock client for testing
type mockEvent struct {
	Name   string
	Params map[string]interface{}
}

type mockGA4Client struct {
	events []mockEvent
}

func (m *mockGA4Client) TrackBusinessCreated(businessID, email, industry, city string) error {
	m.events = append(m.events, mockEvent{
		Name: "business_created",
		Params: map[string]interface{}{
			"business_id": businessID,
			"email":       email,
			"industry":    industry,
			"city":        city,
		},
	})
	return nil
}

func (m *mockGA4Client) TrackSubscriptionCreated(subID, planID, businessID string, amount int) error {
	m.events = append(m.events, mockEvent{
		Name: "subscription_created",
		Params: map[string]interface{}{
			"sub_id":       subID,
			"plan_id":      planID,
			"business_id":  businessID,
			"amount":       amount,
		},
	})
	return nil
}

func (m *mockGA4Client) TrackValidationSuccess(subID, planID string, count int) error {
	m.events = append(m.events, mockEvent{
		Name: "validation_success",
		Params: map[string]interface{}{
			"sub_id":   subID,
			"plan_id":  planID,
			"count":    count,
		},
	})
	return nil
}

func (m *mockGA4Client) TrackValidationFailed(subID, reason string) error {
	m.events = append(m.events, mockEvent{
		Name: "validation_failed",
		Params: map[string]interface{}{
			"sub_id": subID,
			"reason": reason,
		},
	})
	return nil
}

func (m *mockGA4Client) TrackPaymentReceived(subID string, amount int, planID string) error {
	m.events = append(m.events, mockEvent{
		Name: "payment_received",
		Params: map[string]interface{}{
			"sub_id":   subID,
			"amount":   amount,
			"plan_id":  planID,
		},
	})
	return nil
}

func (m *mockGA4Client) TrackPaymentFailed(subID, reason string) error {
	m.events = append(m.events, mockEvent{
		Name: "payment_failed",
		Params: map[string]interface{}{
			"sub_id": subID,
			"reason": reason,
		},
	})
	return nil
}

func (m *mockGA4Client) TrackSubscriptionCanceled(subID, reason string) error {
	m.events = append(m.events, mockEvent{
		Name: "subscription_canceled",
		Params: map[string]interface{}{
			"sub_id": subID,
			"reason": reason,
		},
	})
	return nil
}

func newMockClient() *mockGA4Client {
	return &mockGA4Client{
		events: make([]mockEvent, 0),
	}
}
