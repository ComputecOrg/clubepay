package provider_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/clubepay/backend/internal/provider"
)

type MockCostProvider struct {
	cost int64
}

func (m *MockCostProvider) GetMonthlyCost(ctx context.Context) (provider.Cost, error) {
	return provider.Cost{
		CostCents:   m.cost,
		Provider:    "mock",
		Description: "Mock provider",
	}, nil
}

func TestAggregator_GetTotalMonthlyCost(t *testing.T) {
	providers := []provider.CostProvider{
		&MockCostProvider{cost: 10000},
		&MockCostProvider{cost: 20000},
		&MockCostProvider{cost: 5000},
	}
	agg := provider.NewAggregator(providers)

	cost, err := agg.GetTotalMonthlyCost(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, int64(35000), cost)
}

func TestAggregator_EmptyProviders(t *testing.T) {
	agg := provider.NewAggregator([]provider.CostProvider{})

	cost, err := agg.GetTotalMonthlyCost(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, int64(0), cost)
}
