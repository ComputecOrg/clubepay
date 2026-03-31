package analytics

// Client is the interface for tracking analytics events
type Client interface {
	TrackBusinessCreated(businessID, email, industry, city string) error
	TrackSubscriptionCreated(subID, planID, businessID string, amount int) error
	TrackValidationSuccess(subID, planID string, count int) error
	TrackValidationFailed(subID, reason string) error
	TrackPaymentReceived(subID string, amount int, planID string) error
	TrackPaymentFailed(subID, reason string) error
	TrackSubscriptionCanceled(subID, reason string) error
}

// NoOp is a no-operation analytics client (useful for testing/disabled analytics)
type NoOp struct{}

func (NoOp) TrackBusinessCreated(businessID, email, industry, city string) error {
	return nil
}

func (NoOp) TrackSubscriptionCreated(subID, planID, businessID string, amount int) error {
	return nil
}

func (NoOp) TrackValidationSuccess(subID, planID string, count int) error {
	return nil
}

func (NoOp) TrackValidationFailed(subID, reason string) error {
	return nil
}

func (NoOp) TrackPaymentReceived(subID string, amount int, planID string) error {
	return nil
}

func (NoOp) TrackPaymentFailed(subID, reason string) error {
	return nil
}

func (NoOp) TrackSubscriptionCanceled(subID, reason string) error {
	return nil
}
