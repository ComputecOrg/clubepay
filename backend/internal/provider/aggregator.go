package provider

import (
	"context"
	"log/slog"
)

// Aggregator collects costs from multiple providers
type Aggregator struct {
	providers []CostProvider
}

// NewAggregator creates a new cost aggregator with given providers
func NewAggregator(providers []CostProvider) *Aggregator {
	return &Aggregator{
		providers: providers,
	}
}

// GetTotalInfrastructureCost aggregates costs from all providers without blocking on individual failures
func (a *Aggregator) GetTotalInfrastructureCost(ctx context.Context) int64 {
	var total int64

	for _, p := range a.providers {
		cost, err := p.GetMonthlyCost(ctx)
		if err != nil {
			slog.Error("provider failed to get cost", "error", err)
			// Continue with other providers (non-blocking)
			continue
		}
		total += cost.CostCents
	}

	return total
}
