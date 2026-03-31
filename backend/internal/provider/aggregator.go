package provider

import (
	"context"
	"log/slog"
)

// Aggregator collects costs from multiple providers
type Aggregator struct {
	providers []CostProvider
}

// NewAggregator creates a cost aggregator
func NewAggregator(providers []CostProvider) *Aggregator {
	return &Aggregator{providers: providers}
}

// GetTotalInfrastructureCost sums costs from all providers
// Non-blocking: if a provider fails, log error but continue with others
func (a *Aggregator) GetTotalInfrastructureCost(ctx context.Context) (int64, error) {
	var total int64

	for _, p := range a.providers {
		cost, err := p.GetMonthlyCost(ctx)
		if err != nil {
			slog.Error("provider failed to get cost", "provider", cost.Provider, "error", err)
			// Continue with other providers (non-blocking)
			continue
		}
		total += cost.CostCents
	}

	return total, nil
}
