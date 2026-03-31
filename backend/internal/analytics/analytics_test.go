package analytics

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventTracker_TrackBusinessCreated(t *testing.T) {
	// Arrange
	tracker := NewMockEventTracker()

	// Act
	err := tracker.TrackBusinessCreated(
		"owner-123",
		"Cafe Express",
		"cafe",
		"São Paulo",
	)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 1, len(tracker.recordedEvents))
	assert.Equal(t, "business_created", tracker.recordedEvents[0].Name)
	assert.Equal(t, "owner-123", tracker.recordedEvents[0].Params["user_id"])
	assert.Equal(t, "cafe", tracker.recordedEvents[0].Params["industry"])
	assert.Equal(t, "São Paulo", tracker.recordedEvents[0].Params["city"])
}

func TestEventTracker_TrackSubscriptionCreated(t *testing.T) {
	// Arrange
	tracker := NewMockEventTracker()

	// Act
	err := tracker.TrackSubscriptionCreated(
		"subscriber-456",
		"plan-789",
		"Premium",
		"",
	)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 1, len(tracker.recordedEvents))
	assert.Equal(t, "subscription_created", tracker.recordedEvents[0].Name)
	assert.Equal(t, "subscriber-456", tracker.recordedEvents[0].Params["user_id"])
	assert.Equal(t, "plan-789", tracker.recordedEvents[0].Params["plan_id"])
	assert.Equal(t, "Premium", tracker.recordedEvents[0].Params["plan_name"])
}

func TestEventTracker_TrackValidationSuccess(t *testing.T) {
	// Arrange
	tracker := NewMockEventTracker()

	// Act
	err := tracker.TrackValidationSuccess(
		"subscriber-456",
		"plan-789",
		5,
	)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 1, len(tracker.recordedEvents))
	assert.Equal(t, "validation_success", tracker.recordedEvents[0].Name)
	assert.Equal(t, "subscriber-456", tracker.recordedEvents[0].Params["user_id"])
	assert.Equal(t, "plan-789", tracker.recordedEvents[0].Params["plan_id"])
	assert.Equal(t, float64(5), tracker.recordedEvents[0].Params["daily_count"])
}

func TestEventTracker_TrackPaymentReceived(t *testing.T) {
	// Arrange
	tracker := NewMockEventTracker()

	// Act
	err := tracker.TrackPaymentReceived(
		"owner-123",
		"subscription-111",
		50000, // 500.00 in cents
		"monthly",
	)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 1, len(tracker.recordedEvents))
	assert.Equal(t, "payment_received", tracker.recordedEvents[0].Name)
	assert.Equal(t, "owner-123", tracker.recordedEvents[0].Params["user_id"])
	assert.Equal(t, float64(500.00), tracker.recordedEvents[0].Params["amount"])
	assert.Equal(t, "monthly", tracker.recordedEvents[0].Params["plan_type"])
}

func TestEventTracker_TrackMultipleEvents(t *testing.T) {
	// Arrange
	tracker := NewMockEventTracker()

	// Act
	_ = tracker.TrackBusinessCreated("owner-1", "Cafe 1", "cafe", "SP")
	_ = tracker.TrackSubscriptionCreated("sub-1", "plan-1", "Basic", "")
	_ = tracker.TrackPaymentReceived("owner-1", "sub-1", 10000, "monthly")

	// Assert
	assert.Equal(t, 3, len(tracker.recordedEvents))
	assert.Equal(t, "business_created", tracker.recordedEvents[0].Name)
	assert.Equal(t, "subscription_created", tracker.recordedEvents[1].Name)
	assert.Equal(t, "payment_received", tracker.recordedEvents[2].Name)
}
