package provider

import "context"

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

// GetTotalMonthlyCost aggregates costs from all providers
func (a *Aggregator) GetTotalMonthlyCost(ctx context.Context) (int64, error) {
	total := int64(0)
	for _, provider := range a.providers {
		cost, err := provider.GetMonthlyCost(ctx)
		if err != nil {
			return 0, err
		}
		total += cost.CostCents
	}
	return total, nil
}
