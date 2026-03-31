package provider

import "context"

// Cost represents a cost from a provider
type Cost struct {
	CostCents   int64
	Provider    string
	Description string
}

// CostProvider defines the interface for cost providers
type CostProvider interface {
	GetMonthlyCost(ctx context.Context) (Cost, error)
}
