package provider_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/clubepay/backend/internal/provider"
)

type mockProvider struct {
	cost int64
}

func (m *mockProvider) GetMonthlyCost(ctx context.Context) (provider.Cost, error) {
	return provider.Cost{CostCents: m.cost, Provider: "mock"}, nil
}

func TestAggregator_GetTotalCost(t *testing.T) {
	providers := []provider.CostProvider{
		&mockProvider{cost: 10000}, // $100
		&mockProvider{cost: 5000},  // $50
		&mockProvider{cost: 2500},  // $25
	}
	agg := provider.NewAggregator(providers)

	ctx := context.Background()
	total, err := agg.GetTotalInfrastructureCost(ctx)

	require.NoError(t, err)
	assert.Equal(t, int64(17500), total) // $175 total
}
