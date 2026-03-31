package provider_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/clubepay/backend/internal/provider"
)

type MockCostProvider struct {
	cost int64
	err  error
}

func (m *MockCostProvider) GetMonthlyCost(ctx context.Context) (provider.Cost, error) {
	if m.err != nil {
		return provider.Cost{}, m.err
	}
	return provider.Cost{
		CostCents:   m.cost,
		Provider:    "mock",
		Description: "Mock provider",
	}, nil
}

func TestAggregator_GetTotalInfrastructureCost(t *testing.T) {
	providers := []provider.CostProvider{
		&MockCostProvider{cost: 10000},
		&MockCostProvider{cost: 20000},
		&MockCostProvider{cost: 5000},
	}
	agg := provider.NewAggregator(providers)

	cost := agg.GetTotalInfrastructureCost(context.Background())
	assert.Equal(t, int64(35000), cost)
}

func TestAggregator_EmptyProviders(t *testing.T) {
	agg := provider.NewAggregator([]provider.CostProvider{})

	cost := agg.GetTotalInfrastructureCost(context.Background())
	assert.Equal(t, int64(0), cost)
}

// TestAggregator_NonBlockingErrors verifies that one provider failure doesn't block others (non-blocking behavior)
func TestAggregator_NonBlockingErrors(t *testing.T) {
	providers := []provider.CostProvider{
		&MockCostProvider{cost: 10000},                                    // Success: 10000
		&MockCostProvider{err: errors.New("provider failed")},            // Error: skipped
		&MockCostProvider{cost: 15000},                                    // Success: 15000
	}
	agg := provider.NewAggregator(providers)

	// Should aggregate successfully despite one provider failing
	cost := agg.GetTotalInfrastructureCost(context.Background())
	assert.Equal(t, int64(25000), cost) // 10000 + 15000 (middle provider error is skipped)
}
