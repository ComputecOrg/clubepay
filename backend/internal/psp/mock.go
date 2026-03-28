// backend/internal/psp/mock.go
package psp

import "context"

type MockPSP struct {
	CreateCustomerFn     func(ctx context.Context, req CreateCustomerRequest) (*Customer, error)
	CreateSubscriptionFn func(ctx context.Context, req CreateSubscriptionRequest) (*Subscription, error)
	CancelSubscriptionFn func(ctx context.Context, subscriptionID string) error
	GetSubscriptionFn    func(ctx context.Context, subscriptionID string) (*Subscription, error)
	GetPaymentsFn        func(ctx context.Context, subscriptionID string) ([]Payment, error)
	ValidateWebhookFn    func(payload []byte, signature string) bool
}

func (m *MockPSP) CreateCustomer(ctx context.Context, req CreateCustomerRequest) (*Customer, error) {
	if m.CreateCustomerFn != nil {
		return m.CreateCustomerFn(ctx, req)
	}
	return &Customer{ID: "cus_mock_123"}, nil
}

func (m *MockPSP) CreateSubscription(ctx context.Context, req CreateSubscriptionRequest) (*Subscription, error) {
	if m.CreateSubscriptionFn != nil {
		return m.CreateSubscriptionFn(ctx, req)
	}
	return &Subscription{ID: "sub_mock_123", Status: "ACTIVE"}, nil
}

func (m *MockPSP) CancelSubscription(ctx context.Context, subscriptionID string) error {
	if m.CancelSubscriptionFn != nil {
		return m.CancelSubscriptionFn(ctx, subscriptionID)
	}
	return nil
}

func (m *MockPSP) GetSubscription(ctx context.Context, subscriptionID string) (*Subscription, error) {
	if m.GetSubscriptionFn != nil {
		return m.GetSubscriptionFn(ctx, subscriptionID)
	}
	return &Subscription{ID: subscriptionID, Status: "ACTIVE"}, nil
}

func (m *MockPSP) GetPayments(ctx context.Context, subscriptionID string) ([]Payment, error) {
	if m.GetPaymentsFn != nil {
		return m.GetPaymentsFn(ctx, subscriptionID)
	}
	return []Payment{{ID: "pay_mock_1", Status: "CONFIRMED"}}, nil
}

func (m *MockPSP) ValidateWebhook(payload []byte, signature string) bool {
	if m.ValidateWebhookFn != nil {
		return m.ValidateWebhookFn(payload, signature)
	}
	return true
}
