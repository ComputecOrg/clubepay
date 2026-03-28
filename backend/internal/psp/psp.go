package psp

import "context"

// PSP defines the payment service provider interface.
type PSP interface {
	CreateCustomer(ctx context.Context, req CreateCustomerRequest) (*Customer, error)
	CreateSubscription(ctx context.Context, req CreateSubscriptionRequest) (*Subscription, error)
	CancelSubscription(ctx context.Context, subscriptionID string) error
	GetSubscription(ctx context.Context, subscriptionID string) (*Subscription, error)
	GetPayments(ctx context.Context, subscriptionID string) ([]Payment, error)
	ValidateWebhook(payload []byte, signature string) bool
}

type CreateCustomerRequest struct {
	Name  string
	Email string
	Phone string
	CPF   string
}

type Customer struct {
	ID string
}

type CreateSubscriptionRequest struct {
	CustomerID  string
	PriceCents  int64
	Description string
	Cycle       string // MONTHLY
}

type Subscription struct {
	ID     string
	Status string
}

type Payment struct {
	ID     string
	Status string // CONFIRMED, PENDING, OVERDUE, REFUNDED
}
