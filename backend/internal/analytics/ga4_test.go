package analytics

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGA4EventTracker(t *testing.T) {
	// Act
	tracker := NewGA4EventTracker("G-XXXXXXXXXX", "api-secret-123", "user-id-456")

	// Assert
	assert.NotNil(t, tracker)
	assert.Equal(t, "G-XXXXXXXXXX", tracker.measurementID)
	assert.Equal(t, "api-secret-123", tracker.apiSecret)
	assert.Equal(t, "user-id-456", tracker.userID)
}

func TestGA4EventTracker_TrackBusinessCreated(t *testing.T) {
	// Arrange
	tracker := NewGA4EventTracker("", "", "user-id")

	// Act
	err := tracker.TrackBusinessCreated("owner-1", "Cafe Express", "cafe", "SP")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 1, len(tracker.eventBuffer))
	assert.Equal(t, "business_created", tracker.eventBuffer[0].Name)
}

func TestGA4EventTracker_Flush_WithoutCredentials(t *testing.T) {
	// Arrange - sem credentials, eventos devem ser descartados
	tracker := NewGA4EventTracker("", "", "user-id")
	err := tracker.TrackBusinessCreated("owner-1", "Cafe", "cafe", "SP")
	require.NoError(t, err)

	// Act
	err = tracker.Flush(context.Background())

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 0, len(tracker.eventBuffer))
}

func TestGA4EventTracker_FlushWithMultipleEvents(t *testing.T) {
	// Arrange
	tracker := NewGA4EventTracker("", "", "user-id")
	_ = tracker.TrackBusinessCreated("owner-1", "Cafe", "cafe", "SP")
	_ = tracker.TrackFirstPlanCreated("owner-1", "plan-1", "Basic", 10000)
	_ = tracker.TrackPaymentReceived("owner-1", "sub-1", 50000, "monthly")

	// Act
	err := tracker.Flush(context.Background())

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 0, len(tracker.eventBuffer))
}

func TestGA4EventTracker_AutoFlushOnBufferFull(t *testing.T) {
	// Arrange - cria tracker sem flush automático por timer
	tracker := NewGA4EventTracker("", "", "user-id")

	// Act - adiciona muitos eventos para atingir limite
	for i := 0; i < 51; i++ {
		_ = tracker.TrackBusinessCreated("owner-1", "Cafe", "cafe", "SP")
	}

	// Assert - buffer deve ter sido limpo após atingir 50
	assert.Equal(t, 1, len(tracker.eventBuffer)) // 51º evento ainda no buffer
}

func TestGA4EventTracker_AllEventTypes(t *testing.T) {
	// Arrange
	tracker := NewGA4EventTracker("", "", "user-123")

	// Act - tracks all event types
	_ = tracker.TrackBusinessCreated("owner-1", "Cafe", "cafe", "SP")
	_ = tracker.TrackFirstPlanCreated("owner-1", "plan-1", "Basic", 10000)
	_ = tracker.TrackFirstSubscriber("owner-1", "plan-1")
	_ = tracker.TrackSubscriptionCanceled("owner-1", "sub-1", "refund requested")
	_ = tracker.TrackSubscriptionCreated("sub-1", "plan-1", "Basic", "REF-123")
	_ = tracker.TrackValidationSuccess("sub-1", "plan-1", 5)
	_ = tracker.TrackValidationFailed("sub-1", "limit exceeded")
	_ = tracker.TrackPlanUpgraded("sub-1", "plan-1", "plan-2")
	_ = tracker.TrackCancellation("sub-1", "too expensive")
	_ = tracker.TrackPaymentReceived("owner-1", "sub-1", 50000, "monthly")
	_ = tracker.TrackPaymentFailed("owner-1", "sub-1", "card declined")

	// Assert
	err := tracker.Flush(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 0, len(tracker.eventBuffer))
}

func TestGA4EventTracker_ParamConversions(t *testing.T) {
	// Arrange
	tracker := NewGA4EventTracker("", "", "user-id")

	// Act
	_ = tracker.TrackPaymentReceived("owner-1", "sub-1", 12500, "monthly") // $125.00
	event := tracker.eventBuffer[0]

	// Assert - verifica conversão de centavos para reais
	amount := event.Params["amount"].(float64)
	assert.Equal(t, 125.0, amount)
}

func TestGA4EventTracker_Close(t *testing.T) {
	// Arrange
	tracker := NewGA4EventTracker("G-XXXXXXXXXX", "secret", "user-id")
	_ = tracker.TrackBusinessCreated("owner-1", "Cafe", "cafe", "SP")

	// Act
	err := tracker.Close(context.Background())

	// Assert
	require.NoError(t, err)
	// Eventos devem ter sido flushed
	assert.Equal(t, 0, len(tracker.eventBuffer))
}
